terraform {
  required_providers {
    coder = {
      source = "coder/coder"
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

data "coder_workspace" "me" {}
resource "coder_agent" "main" {
  arch = "amd64"
  os   = "linux"
}
