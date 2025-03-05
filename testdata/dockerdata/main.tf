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

data "coder_parameter" "example" {
  name        = "Example"
  description = "An example parameter that has no purpose."
  type        = "string"
  default     = trimprefix(data.docker_registry_image.ubuntu.sha256_digest, "sha256:")

  option {
    name = "Ubuntu"
    description = data.docker_registry_image.ubuntu.name
    value = data.docker_registry_image.ubuntu.sha256_digest
  }

  option {
    name = "Centos"
    description = data.docker_registry_image.centos.name
    value = data.docker_registry_image.centos.sha256_digest
  }
}

data "coder_workspace_tags" "custom_workspace_tags" {
  tags = {
    "foo" = data.docker_registry_image.ubuntu.sha256_digest
    "bar" = data.docker_registry_image.centos.sha256_digest
    "qux" = "quux"
  }
}

# Pulls the image
data "docker_registry_image" "centos" {
  name = "centos:centos7.9.2009"
}

data "docker_registry_image" "ubuntu" {
  name = "ubuntu:24.04"
  // sha256_digest
}

