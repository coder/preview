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

data "coder_workspace_tags" "custom_workspace_tags" {
  tags = {
    // If a value is required, you can do something like:
    // try(docker_image.ubuntu.repo_digest, "default-value")
    "foo" = try(docker_image.ubuntu.repo_digest, "default")
    "bar" = docker_image.centos.repo_digest
    "qux" = "quux"
  }
}


# Pulls the image
resource "docker_image" "ubuntu" {
  name = "ubuntu:latest"
}

resource "docker_image" "centos" {
  name = "centos:latest"
}

data "docker_registry_image" "ubuntu" {
  name = "ubuntu:precise"
  // sha256_digest
}

