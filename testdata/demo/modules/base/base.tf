terraform {
  required_providers {
    coder = {
      source = "coder/coder"
      version = "2.4.0-pre0"
    }
  }
}

locals {
  // default to the only option if only 1 exists
  choose_security = length(keys(module.deploys.security_levels)) > 1
  secutity_level = local.choose_security ? data.coder_parameter.security_level[0].value : keys(module.deploys.security_levels)[0]
}

module "deploys" {
  // Where we deploy the workspace to
  source = "../deploys"
  security = local.secutity_level
}

data "coder_parameter" "security_level" {
    count        = local.choose_security ? 1 : 0
    name         = "security_level"
    display_name = "Security Level"
    description  = "What security level do you need?"
    type         = "string"
    default      = "high"
    order        = 50


    dynamic "option" {
        for_each = module.deploys.security_levels
        content {
          name  = option.value.display_name
          value = option.key
          description = option.value.description
        }
    }

    # validation {
    #   regex = "^high|medium|low$"
    #   error = "You must select either high, medium, or low."
    # }
}


output "security_level" {
  value = local.secutity_level
}