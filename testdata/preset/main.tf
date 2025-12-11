terraform {
  required_providers {
    coder = {
      source = "coder/coder"
      version = "2.8.0"
    }
  }
}

data "coder_parameter" "number" {
  name = "number"
  default = "7"
  type = "number"

  validation {
    min = 4
    max = 10
  }
}

data "coder_parameter" "has_default" {
  name = "has_default"
  default = "hello world"
}


data "coder_workspace_preset" "valid_preset" {
  name = "valid_preset"

  parameters = {
    "number" = "9"
    "has_default" = "changed"
  }
  prebuilds {
    instances = 3
  }
}

data "coder_workspace_preset" "prebuild_instance_zero" {
  name = "prebuild_instance_zero"

  prebuilds {
    // No instances
    instances = 0
  }
}

data "coder_workspace_preset" "not_prebuild" {
  name = "not_prebuild"
}