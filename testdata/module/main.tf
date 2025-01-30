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
  count          = data.coder_workspace.me.start_count
  source         = "registry.coder.com/modules/jetbrains-gateway/coder"
  version        = "1.0.27"
  agent_name = "main"
  agent_id       = coder_agent.main.id
  folder         = "/home/coder/example"
  jetbrains_ides = ["CL", "GO", "IU", "PY", "WS"]
  default        = "GO"
}

data "coder_workspace" "me" {}
resource "coder_agent" "main" {
  arch = "amd64"
  os   = "linux"
}


data "coder_workspace_tags" "custom_workspace_tags" {
  tags = {
    "foo" = data.docker_registry_image.ubuntu.sha256_digest
  }
}


data "docker_registry_image" "ubuntu" {
  name = "ubuntu:precise"
  // sha256_digest
}