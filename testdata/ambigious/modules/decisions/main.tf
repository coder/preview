terraform {
  required_providers {
    coder = {
      source = "coder/coder"
    }
  }
}

data coder_workspace_owner "me" {}

output "isAdmin" {
  value = contains(data.coder_workspace_owner.me.groups, "admin")
}

output "groups" {
  value = data.coder_workspace_owner.me.groups
}