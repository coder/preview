// Exercises target-driven resource-closure pruning end to end: parameters and
// workspace tags whose values flow from resource blocks (directly, through a
// local, and through a count index) must be identical whether or not the
// unreferenced resources are pruned. The orphan resource below is outside the
// closure and is dropped; it must not change any output.
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
  name  = "poolimg"
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

data "coder_workspace_tags" "tags" {
  tags = {
    flavor = docker_image.base.name
  }
}

// Nothing in the parameter/tag closure references this resource, so it is
// pruned. Its presence must not affect the parameters or tags above.
resource "docker_container" "orphan" {
  name  = "orphan"
  image = "does-not-exist"
}
