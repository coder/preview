terraform {
  required_providers {
    coder = {
      source  = "coder/coder"
      version = "2.4.0-pre0"
    }
  }
}

data "coder_parameter" "region" {
  name    = "region"
  type    = "string"
  default = "us"

  option {
    name  = "US"
    value = "us"
  }
  option {
    name  = "EU"
    value = "eu"
  }
}

data "coder_parameter" "size" {
  name    = "size"
  type    = "number"
  # Invalid value should become valid once the options are overridden.
  default = 100

  option {
    name  = "10GB"
    value = 10
  }
  option {
    name  = "20GB"
    value = 20
  }
}

data "coder_workspace_preset" "dev" {
  name = "dev"
  parameters = {
    region = "us"
  }
}

data "coder_workspace_tags" "tags" {
  tags = {
    "env" = "staging"
  }
}

variable "string_to_number" {
  type    = string
  default = "foo"
}
