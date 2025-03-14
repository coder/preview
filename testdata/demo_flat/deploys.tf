// Handles which cluster the workspace should be deployed to
// using workspace tags.
locals {
  security_levels = {
    "high"   = {
      display_name = "High"
      description  = "Most confidentiality, restricted access. Deployed into the confidential cluster."
      tags         = {"cluster": "confidential"}
    }
    "medium" = {
      display_name = "Medium"
      description  = "A medium security level. Deployed into the standard production cluster."
      tags         = {"cluster": "production"}
    }
    "low"    = {
      display_name = "Low"
      description  = "The lowest security level. Deployed into the public cluster."
      tags         = {"cluster": "public"}
    }
  }

  admin = local.security_levels
  developer = {for k in ["high", "medium"] : k => local.security_levels[k]}
  contractor = {for k in ["high"] : k => local.security_levels[k]}
  isAdmin = contains(data.coder_workspace_owner.me.groups, "admin")
  isDeveloper = contains(data.coder_workspace_owner.me.groups, "developer")

  allowed_security_levels = (
    local.isAdmin ? local.admin :
      local.isDeveloper ? local.developer : local.contractor
  )

  direct_ssh_allowed = local.isAdmin && var.security == "low"
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
