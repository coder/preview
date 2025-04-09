terraform {
  required_providers {
    coder = {
      version = "v2.4.0-pre0"
      source = "coder/coder"
    }
  }
}

data "coder_parameter" "project" {
  name        = "Project"
  description = "Which project are you working on?"
  type        = "string"
  default     = "massive"
  order       = 1

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

  validation {
    regex = "^massive|small$"
    error = "You must select either massive or small."
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

  use_options = data.coder_parameter.project.value == "massive" ?  local.massive_options : local.small_options
}

data "coder_parameter" "compute" {
  name        = "Compute"
  description = "How much compute do you need?"
  type        = "string"
  default     = data.coder_parameter.project.value == "massive" ? "huge" : "small"
  order       = 2

  dynamic "option" {
    for_each = local.use_options
    content {
      name = option.value.name
      value = option.value.value
    }
  }
}

data coder_workspace_owner "me" {}
locals {
  isAdmin = contains(data.coder_workspace_owner.me.groups, "admin")
}

data "coder_parameter" "image_hash" {
  count       = local.isAdmin ? 1 : 0
  name = "hash"
  display_name        = "Image Hash"
  description = "Override the hash of the image to use. Only available to admins."
  // Value can get stale
  default     = "e64c69d84d5f910b5cd4fc7bc01a67a6436865787b429e7e60ebaeb4e7dd1b44"
  order       = 3

  validation {
    regex = "^[a-f0-9A-F]{64}$"
    error = "The image hash must be a 64-character hexadecimal string."
  }
}