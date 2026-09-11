// Second batch of resource-reachability vectors (see resourceclosure for the
// first). Each parameter, preset, or tag below reads a resource through an
// evaluation path that is distinct from the first batch: validation blocks,
// for expressions, a splat outside a dynamic block, a count = 0 resource, a
// resource whose for_each is driven by a parameter, a computed attribute, a
// two-level module chain, and multi-hop locals. Any strategy that skips
// resources must leave every value here unchanged.
terraform {
  required_providers {
    coder = {
      source  = "coder/coder"
      version = "2.4.0-pre0"
    }
    docker = {
      source  = "kreuzwerker/docker"
      version = "3.0.2"
    }
  }
}

resource "docker_image" "base" {
  name = "large"
}

resource "docker_image" "pool" {
  count = 2
  name  = "pool-${count.index}"
}

resource "docker_image" "bykey" {
  for_each = toset(["a", "b"])
  name     = "img-${each.key}"
}

// --- validation block reads a resource -------------------------------------
data "coder_parameter" "validated" {
  name    = "validated"
  type    = "string"
  default = docker_image.base.name
  validation {
    regex = "^${docker_image.base.name}$"
    error = "must be ${docker_image.base.name}"
  }
}

// --- for expressions over resources ----------------------------------------
locals {
  pool_names = [for i in docker_image.pool : i.name]
  bykey_map  = { for k, v in docker_image.bykey : k => v.name }
}

data "coder_parameter" "forlist" {
  name    = "forlist"
  type    = "string"
  default = join(",", local.pool_names)
}

data "coder_parameter" "formap" {
  name    = "formap"
  type    = "string"
  default = local.bykey_map["b"]
}

// --- splat directly in a tag ------------------------------------------------
data "coder_workspace_tags" "tags" {
  tags = {
    pool = join(",", docker_image.pool[*].name)
    // A resource whose for_each is driven by a parameter (below). It both
    // reads a parameter and feeds a target.
    shards = tostring(length(docker_image.shard))
  }
}

// --- count = 0 resource: kept, expands to nothing ---------------------------
resource "docker_image" "none" {
  count = 0
  name  = "never"
}

data "coder_parameter" "fallback" {
  name    = "fallback"
  type    = "string"
  default = try(docker_image.none[0].name, "fallback")
}

// --- resource for_each driven by a parameter value --------------------------
data "coder_parameter" "shards" {
  name    = "shards"
  type    = "list(string)"
  default = jsonencode(["x", "y", "z"])
}

resource "docker_image" "shard" {
  for_each = toset(jsondecode(data.coder_parameter.shards.value))
  name     = "shard-${each.key}"
}

// --- computed attribute: unknown with or without the resource ---------------
data "coder_parameter" "computed" {
  name    = "computed"
  type    = "string"
  default = docker_image.base.image_id
}

// --- two-level module chain --------------------------------------------------
module "outer" {
  source     = "./outer"
  image_name = docker_image.base.name
}

data "coder_parameter" "nested" {
  name    = "nested"
  type    = "string"
  default = module.outer.inner_name
}

// --- multi-hop locals ---------------------------------------------------------
locals {
  hop_c = docker_image.base.name
  hop_b = local.hop_c
  hop_a = local.hop_b
}

data "coder_parameter" "hops" {
  name    = "hops"
  type    = "string"
  default = local.hop_a
}

// --- orphan: reachable from nothing -----------------------------------------
resource "docker_container" "orphan" {
  name  = "orphan"
  image = "does-not-exist"
}
