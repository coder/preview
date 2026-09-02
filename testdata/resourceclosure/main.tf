// Exercises target-driven resource-closure pruning end to end. Every resource
// below is reachable from exactly one target (a parameter, a preset, or a tag)
// or through a resource->resource chain, so each proves that its path keeps the
// resource alive through pruning. The orphan is reachable from nothing and is
// dropped; it must not change any output. Resource attributes use static values
// so the assertions isolate pruning behaviour, not count/for_each value binding.
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

// --- Parameter-reached resources -------------------------------------------
resource "docker_image" "base" {
  name = "large"
}

resource "docker_image" "pool" {
  count = 2
  name  = "poolimg"
}

resource "docker_image" "bykey" {
  for_each = toset(["a"])
  name     = "fe-large"
}

// Transitive chain: a parameter reaches chain_a directly, chain_a reaches
// chain_b. Pruning must keep both.
resource "docker_image" "chain_b" {
  name = "chain-large"
}

resource "docker_image" "chain_a" {
  name = docker_image.chain_b.name
}

locals {
  flavor = docker_image.base.name
}

data "coder_parameter" "flavor" {
  name    = "flavor"
  type    = "string"
  default = local.flavor
}

data "coder_parameter" "direct" {
  name    = "direct"
  type    = "string"
  default = docker_image.base.name
}

data "coder_parameter" "indexed" {
  name    = "indexed"
  type    = "string"
  default = docker_image.pool[0].name
}

data "coder_parameter" "byeach" {
  name    = "byeach"
  type    = "string"
  default = docker_image.bykey["a"].name
}

data "coder_parameter" "chained" {
  name    = "chained"
  type    = "string"
  default = docker_image.chain_a.name
}

// --- Preset-reached resource (nothing else references it) ------------------
resource "docker_image" "preset_only" {
  name = "preset-large"
}

data "coder_workspace_preset" "big" {
  name = "big"
  parameters = {
    flavor = docker_image.preset_only.name
  }
}

// --- Tag-reached resource (nothing else references it) ---------------------
resource "docker_image" "tag_only" {
  name = "tag-large"
}

data "coder_workspace_tags" "tags" {
  tags = {
    flavor = docker_image.base.name
    tagged = docker_image.tag_only.name
  }
}

// --- Orphan: reachable from no target, pruned, must not affect outputs -----
resource "docker_container" "orphan" {
  name  = "orphan"
  image = "does-not-exist"
}
