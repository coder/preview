terraform {
  required_providers {
    coder = {
      source = "coder/coder"
      version = "v2.4.0-pre0"
    }
  }
}

data coder_workspace_owner "me" {}

module "decisions" {
  source = "./modules/decisions"
}

data "coder_parameter" "zero" {
  count       =  module.decisions.zero
  name        = "Zero"
  type        = "bool"
  default     = false
}