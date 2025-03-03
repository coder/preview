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