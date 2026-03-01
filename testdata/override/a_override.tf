# Override region's default (will be overridden again by b_override).
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
    name  = "50GB"
    value = 50
  }
}

# Override size again in the same file — adds 100GB option that makes the
# default valid.
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
    "env" = "production"
  }
}

# Override variable.
variable "string_to_number" {
  type = number
  default = 40
}
