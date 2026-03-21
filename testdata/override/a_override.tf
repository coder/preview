terraform {
  # An override (will be skipped).
  required_version = ">= 1.1"
}

# Override region's default (will be overridden again by override.tf).
data "coder_parameter" "region" {
  default = "eu"
}

# Override size's options.
data "coder_parameter" "size" {
  option {
    name  = "10GB"
    value = 10
  }
  option {
    name  = "40GB"
    value = 40
  }
}

locals {
  # Override the local value in main.tf.
  default_size = 50
}

# Override size again in the same file — adds 50GB option that makes the
# overridden default valid.
data "coder_parameter" "size" {
  option {
    name  = "10GB"
    value = 10
  }
  option {
    name  = "50GB"
    value = 50
  }
  option {
    name  = "100GB"
    value = 100
  }
}

# Override tags.
data "coder_workspace_tags" "tags" {
  tags = {
    "env"  = "production"
    "team" = "mango"
  }
}

# Override static options with dynamic.
data "coder_parameter" "static_to_dynamic" {
  dynamic "option" {
    for_each = var.zones
    content {
      name  = option.value
      value = option.value
    }
  }
}

# Override dynamic options with static.
data "coder_parameter" "dynamic_to_static" {
  option {
    name  = "X"
    value = "x"
  }
  option {
    name  = "Y"
    value = "y"
  }
}

# Override variable.
variable "string_to_number" {
  type = number
  default = 40
}
