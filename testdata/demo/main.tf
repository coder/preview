// Demo terraform has a complex configuration.
// CODER_WORKSPACE_OWNER_GROUPS='["admin","developer"]' terraform apply
//
// Some run options
// preview -v Team=backend -g admin
// preview -v Team=backend -g admin -g sa-saopaulo
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


data coder_workspace_owner "me" {}

module "jetbrains_gateway" {
  count          = 1
  source         = "registry.coder.com/modules/jetbrains-gateway/coder"
  version        = "1.0.28"
  agent_id       = "random"
  folder         = "/home/coder/example"
  jetbrains_ides = local.teams[data.coder_parameter.team.value].codes
  default        = local.teams[data.coder_parameter.team.value].codes[0]
  coder_parameter_order = 11
}

module "base" {
  source = "./modules/base"
}