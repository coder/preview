terraform {
  required_providers {
    coder = {
      source = "coder/coder"
    }
  }
}

data "coder_parameter" "project" {
  name        = "Project"
  description = "Which project are you working on?"
  type        = "string"
  default     = "massive"

  option {
    name  = "Massive Project"
    value = "massive"
    description = "The massive project with a ton of resource requirements to work on."
  }
  option {
    name  = "Small Project"
    value = "small"
    description = "The small project with minimal resource requirements to work on."
  }
}

locals {
  small_options = [
    {
      name  = "Micro"
      value = "micro"
    },
    {
      name  = "Small"
      value = "small"
    },
  ]

  massive_options = concat(local.small_options, [
    {
    name  = "Medium"
    value = "medium"
    },
    {
    name  = "Huge"
    value = "huge"
    },
  ])

  use_options = data.coder_parameter.project.default == "massive" ?  local.massive_options : local.small_options
}

data "coder_parameter" "compute" {
  name        = "Compute"
  description = "How much compute do you need?"
  type        = "string"
  default     = data.coder_parameter.project.value == "massive" ? "huge" : "small"
  validation {

  }

  dynamic "option" {
    for_each = local.use_options
    content {
      name = option.value.name
      value = option.value.value
    }
  }
}