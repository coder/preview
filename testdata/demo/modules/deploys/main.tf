// Handles which cluster the workspace should be deployed to
// using workspace tags.
terraform {
  required_providers {
    coder = {
      source = "coder/coder"
      version = "2.4.0-pre0"
    }
  }
}

data coder_workspace_owner "me" {}

output "security_levels" {
  // Output the security levels available to the user
  value = local.allowed_security_levels
}

variable "security" {
  type = string
  default = "high"
}

data "coder_workspace_tags" "custom_workspace_tags" {
  tags = local.security_levels[var.security].tags
}


data "coder_parameter" "direct_ssh" {
    count       =  local.direct_ssh_allowed ? 1 : 0
    name        = "Direct SSH to Pod"
    description = "Should direct SSH access be enabled to the workspace pod? This should be set to false for production workspaces, and is a debugging tool."
    type        = "bool"
    default     = false
}
