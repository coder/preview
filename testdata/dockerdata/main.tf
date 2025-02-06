# main.tf

terraform {
  required_providers {
    coder = {
      source = "coder/coder"
    }
    docker = {
      source  = "kreuzwerker/docker"
      version = "3.0.2"
    }
  }
}

module "jetbrains_gateway" {
  count          = 1
  source         = "registry.coder.com/modules/jetbrains-gateway/coder"
  version        = "1.0.28"
  agent_id       = data.coder_parameter.example.id
  folder         = "/home/coder/example"
  jetbrains_ides = ["CL", "GO", "IU", "PY", "WS"]
  default        = "GO"
}

data "coder_parameter" "example" {
  name        = "Example"
  description = "An example parameter that has no purpose."
  type        = "string"

  option {
    name = "Ubuntu"
    description = data.docker_registry_image.ubuntu.name
    value = try(data.docker_registry_image.ubuntu.sha256_digest, "??")
  }

  option {
    name = "Centos"
    description = docker_image.centos.name
    value = try(docker_image.centos.repo_digest, "??")
  }
}

data "coder_workspace_tags" "custom_workspace_tags" {
  tags = {
    // If a value is required, you can do something like:
    // try(docker_image.ubuntu.repo_digest, "default-value")
    "foo" = data.docker_registry_image.ubuntu.sha256_digest
    "bar" = docker_image.centos.repo_digest
    "qux" = "quux"
  }
}

# Pulls the image
resource "docker_image" "centos" {
  name = "centos:latest"
}

data "docker_registry_image" "ubuntu" {
  name = "ubuntu:precise"
  // sha256_digest
}

