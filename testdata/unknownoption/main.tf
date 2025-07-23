# main.tf

terraform {
  required_providers {
    coder = {
      source = "coder/coder"
      version = "2.4.0-pre0"
    }
    docker = {
      source  = "kreuzwerker/docker"
      version = "3.0.2"
    }
  }
}

variable "unknown" {
  default = null
}

data "coder_parameter" "unknown" {
  name = "unknown"
  display_name = "Unknown Option Example"
  type        = "string"
  default = "foo"

  option {
    name = "Ubuntu"
    value = data.docker_registry_image.ubuntu.sha256_digest
  }

  option {
    name  = "foo"
    value = "foo"
  }
}

data "docker_registry_image" "ubuntu" {
  name = "ubuntu:24.04"
  // sha256_digest
}

