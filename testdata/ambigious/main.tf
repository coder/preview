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

data "coder_parameter" "example" {
  count       =  module.decisions.isAdmin ? 1 : 0
  name        = "IsAdmin"
  type        = "bool"
  default     =  module.decisions.isAdmin
}

data "coder_parameter" "never-show" {
  count       =  module.decisions.staticFalse ? 1 : 0
  name        = "NeverShow"
  type        = "bool"
  default     =  module.decisions.staticFalse
}

data "coder_parameter" "example_root" {
  count       =  contains(data.coder_workspace_owner.me.groups, "admin") ? 1 : 0
  name        = "IsAdmin_Root"
  type        = "bool"
  default     =  contains(data.coder_workspace_owner.me.groups, "admin")
}

output "groups" {
  value = module.decisions.groups
}

output "isAdmin" {
  value = module.decisions.isAdmin
}