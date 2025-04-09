terraform {
  required_providers {
    coder = {
      source = "coder/coder"
      version = "v2.4.0-pre0"
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
  version        = "1.0.27"
  agent_name = "main"
  agent_id       = coder_agent.main.id
  folder         = "/home/coder/example"
  jetbrains_ides = ["CL", "GO", "IU", "PY", "WS"]
  default        = "GO"
}

data "coder_parameter" "extra" {
  count = 1
  name = "extra"
  display_name = "Extra Param"
  description = "A param to throw into the mix."
  type        = "string"
  default     = trimprefix(data.docker_registry_image.coder[1].sha256_digest, "sha256:")
}

data "coder_workspace" "me" {}
resource "coder_agent" "main" {
  arch = "amd64"
  os   = "linux"
}

data "docker_registry_image" "coder" {
  count = 2
  name = count.index == 0 ? "ghcr.io/coder/coder:latest": "ghcr.io/coder/coder:v2.20.1"
}