// Demo terraform has a complex configuration.
terraform {
  required_providers {
    coder = {
      source = "coder/coder"
    }
  }
}


locals {
  fe_codes = ["PS", "WS"]
  be_codes = ["CL", "GO", "IU", "PY"]
  teams = {
    "frontend" = {
        "display_name" = "Frontend",
        "codes" = local.fe_codes,
        "description" = "The team that works on the frontend.",
        "icon" = "/icon/desktop.svg"
    },
    "backend" = {
        "display_name" = "Backend",
        "codes" = local.be_codes,
        "description" = "The team that works on the backend.",
        "icon" = "/emojis/2699.png",
    },
    "fullstack" = {
        "display_name" = "Fullstack",
        "codes" = concat(local.be_codes, local.fe_codes),
        "description" = "The team that works on both the frontend and backend.",
        "icon" = "/emojis/1f916.png",
    }
  }
}

data "coder_parameter" "team" {
    name        = "Team"
    description = "Which team are you on?"
    type        = "string"
    default     = "fullstack"
    order       = 1

    dynamic "option" {
        for_each = local.teams
        content {
            name  = option.value.display_name
            value = option.key
            description = option.value.description
            icon = option.value.icon
        }
    }

    validation {
        regex = "^frontend|backend|fullstack$"
        error = "You must select either frontend, backend, or fullstack."
    }
}

module "jetbrains_gateway" {
  count          = 1
  source         = "registry.coder.com/modules/jetbrains-gateway/coder"
  version        = "1.0.28"
  agent_id       = "random"
  folder         = "/home/coder/example"
  jetbrains_ides = local.teams[data.coder_parameter.team.value].codes
  default        = local.teams[data.coder_parameter.team.value].codes[0]
}

module "base" {
  source = "./modules/base"
}