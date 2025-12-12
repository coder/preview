terraform {
  required_providers {
    coder = {
      source = "coder/coder"
      version = "2.8.0"
    }
  }
}

data "coder_parameter" "valid_parameter" {
  name = "valid_parameter_name"
  default = "valid_option_value"
  option {
    name = "valid_option_name"
    value = "valid_option_value"
  }
}

data "coder_workspace_preset" "no_parameters" {
  name = "no_parameters"
}

data "coder_workspace_preset" "empty_parameters" {
  name = "empty_parameters"
  parameters = {}
}

data "coder_workspace_preset" "invalid_parameter_name" {
  name = "invalid_parameter_name"
  parameters = {
    "invalid_parameter_name" = "irrelevant_value"
  }
}

data "coder_workspace_preset" "invalid_parameter_value" {
  name = "invalid_parameter_value"
  parameters = {
    "valid_parameter_name" = "invalid_value"
  }
}

data "coder_workspace_preset" "valid_preset" {
  name = "valid_preset"
  parameters = {
    "valid_parameter_name" = "valid_option_value"
  }
}

data "coder_workspace_preset" "another_default_preset" {
  name = "another_default_preset"
  parameters = {
    "valid_parameter_name" = "valid_option_value"
  }
  default = true
}

data "coder_workspace_preset" "default_preset" {
  name = "default_preset"
  parameters = {
    "valid_parameter_name" = "valid_option_value"
  }
  default = true
}

