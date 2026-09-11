// The only parameter lives in a submodule and reads a root resource through a
// module input. The root module has no parameter, preset, or tag block of its
// own, so a strategy that prunes root resources based on root targets has no
// basis to prune here and must keep everything.
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

module "sub" {
  source     = "./sub"
  image_name = docker_image.base.name
}

resource "docker_container" "orphan" {
  name  = "orphan"
  image = "does-not-exist"
}
