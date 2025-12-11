terraform {
  required_providers {
    coder = {
      source = "coder/coder"
      version = "2.8.0"
    }
  }
}

data "coder_parameter" "no_default" {
  name = "no_default"
}

data "coder_parameter" "has_default" {
  name = "has_default"
  default = "hello world"
}


data "coder_workspace_preset" "invalid_parameters" {
  name = "invalid_parameters"

  prebuilds {
    instances = 1
  }
}

data "coder_workspace_preset" "valid_preset" {
  name = "valid_preset"

  parameters = {
      "no_default" = "custom value"
      "has_default" = "changed"
  }
  prebuilds {
    instances = 1
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