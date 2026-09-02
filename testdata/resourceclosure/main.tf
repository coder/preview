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

// --- Reference shapes beyond a plain attribute traversal ------------------
// Each of these reaches a resource through an expression form the pruner has
// to see through. If any is missed, the parameter default goes unknown.

// Conditional and try(): both operands are references.
variable "pick_alt" {
  type    = bool
  default = true
}

resource "docker_image" "alt" {
  name = "alt-large"
}

data "coder_parameter" "picked" {
  name    = "picked"
  type    = "string"
  default = var.pick_alt ? docker_image.alt.name : docker_image.base.name
}

data "coder_parameter" "tried" {
  name    = "tried"
  type    = "string"
  default = try(docker_image.alt.name, "fallback")
}

// Index that is itself an expression, not a literal.
variable "idx" {
  type    = number
  default = 1
}

data "coder_parameter" "varindexed" {
  name    = "varindexed"
  type    = "string"
  default = docker_image.pool[var.idx].name
}

// --- References inside nested blocks ----------------------------------------
// A static option block and a dynamic option block both reach a resource.
data "coder_parameter" "staticopt" {
  name      = "staticopt"
  type      = "string"
  form_type = "dropdown"
  default   = docker_image.base.name
  option {
    name  = docker_image.base.name
    value = docker_image.base.name
  }
}

data "coder_parameter" "dynopts" {
  name      = "dynopts"
  type      = "string"
  form_type = "dropdown"
  default   = "poolimg"
  dynamic "option" {
    for_each = toset(docker_image.pool[*].name)
    content {
      name  = option.value
      value = option.value
    }
  }
}

// --- Meta-argument on a target reads a resource -----------------------------
resource "docker_image" "gate" {
  name = "gate"
}

data "coder_parameter" "gated" {
  count   = docker_image.gate.name == "gate" ? 1 : 0
  name    = "gated"
  type    = "string"
  default = "on"
}

// --- Data source as intermediary: param -> data -> resource -----------------
data "docker_registry_image" "viaresource" {
  name = docker_image.base.name
}

data "coder_parameter" "viadata" {
  name    = "viadata"
  type    = "string"
  default = data.docker_registry_image.viaresource.name
}

// --- Tag -> resource -> parameter -------------------------------------------
// The common direction (resource reads a parameter) combined with the resource
// also being in a target's closure. It must be kept and evaluate after the
// parameter it reads.
resource "docker_image" "fromparam" {
  name = "derived-${data.coder_parameter.flavor.value}"
}

data "coder_workspace_tags" "derived" {
  tags = {
    derived = docker_image.fromparam.name
  }
}

// --- Preset nested block reads a resource -----------------------------------
data "coder_workspace_preset" "pre" {
  name = "pre"
  parameters = {
    flavor = "large"
  }
  prebuilds {
    instances = length(docker_image.pool)
  }
}

// --- Module: root resource flows in through an input; the module's own
// resource is untouched by pruning -------------------------------------------
module "sub" {
  source     = "./submodule"
  image_name = docker_image.base.name
}

data "coder_parameter" "viamodule" {
  name    = "viamodule"
  type    = "string"
  default = module.sub.image_name
}

data "coder_parameter" "modresource" {
  name    = "modresource"
  type    = "string"
  default = module.sub.internal_name
}
