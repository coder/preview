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


data "coder_parameter" "hash" {
  # count       = 1
  name        = "hash"
  display_name = "Hash"
  description = "The hash of the image"
  type        = "string"
  default     = trimprefix(data.docker_registry_image.coder.sha256_digest, "sha256:")
}

data "docker_registry_image" "coder" {
  name = "ghcr.io/coder/coder:latest"
}
