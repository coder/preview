terraform {
  required_providers {
    coder = {
      source  = "coder/coder"
      version = "2.4.0-pre0"
    }
  }
}

terraform {
  required_version = ">= 1.0"
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

locals {
  # Just so that we have >1 locals blocks.
  foo = "bar"
}

locals {
  # Will be overridden in a_override.tf
  default_size = 70
}

data "coder_parameter" "size" {
  name    = "size"
  type    = "number"
  # Invalid value should become valid once the options and locals are overridden.
  default = local.default_size

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
    "env"  = "staging"
  }
}

variable "zones" {
  type    = set(string)
  default = ["a", "b", "c"]
}

# Static options, will be overridden by dynamic "option" in a_override.
data "coder_parameter" "static_to_dynamic" {
  name    = "static_to_dynamic"
  type    = "string"
  default = "a"

  option {
    name  = "A"
    value = "a"
  }
  option {
    name  = "B"
    value = "b"
  }
}

# Dynamic options, will be overridden by static option blocks in a_override.
data "coder_parameter" "dynamic_to_static" {
  name    = "dynamic_to_static"
  type    = "string"
  # Invalid value should become valid once the options are overridden.
  default = "x"

  dynamic "option" {
    for_each = var.zones
    content {
      name  = option.value
      value = option.value
    }
  }
}

variable "string_to_number" {
  type    = string
  default = "foo"
}
