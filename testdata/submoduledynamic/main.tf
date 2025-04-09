terraform {
  required_providers {
    coder = {
      source = "coder/coder"
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