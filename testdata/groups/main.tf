terraform {
  required_providers {
    coder = {
      source = "coder/coder"
    }
  }
}

data coder_workspace_owner "me" {}

output "groups" {
  value = data.coder_workspace_owner.me.groups
}

data "coder_parameter" "groups" {
  name = "groups"
  dynamic "option" {
    for_each = data.coder_workspace_owner.me.groups
    content {
      name  = option.value
      value = option.value
    }
  }
}