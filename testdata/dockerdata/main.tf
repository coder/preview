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

locals {
  empty_hash  = "0000000000000000000000000000000000000000000000000000000000000000"
  ubuntu_hash = try(trimprefix(data.docker_registry_image.ubuntu.sha256_digest, "sha256:"), local.empty_hash)
  centos_hash = try(trimprefix(data.docker_registry_image.centos.sha256_digest, "sha256:"), local.empty_hash)
}

data "coder_parameter" "os" {
  name = "os"
  display_name = "Choose your operating system"
  description = "An example parameter that has no purpose."
  type        = "string"
  default     = local.ubuntu_hash

  option {
    name = "Ubuntu (${substr(local.ubuntu_hash, 0, 6)}...${substr(local.ubuntu_hash, 58, 64)})"
    description = data.docker_registry_image.ubuntu.name
    value = local.ubuntu_hash
  }

  option {
    name = "Centos (${substr(local.centos_hash, 0, 6)}...${substr(local.centos_hash, 58, 64)})"
    description = data.docker_registry_image.centos.name
    value = local.centos_hash
  }
}

data "coder_workspace_tags" "custom_workspace_tags" {
  tags = {
    "ubuntu" = local.ubuntu_hash
    "centos" = local.centos_hash
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

