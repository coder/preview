// Base case for workspace tags + parameters.
terraform {
  required_providers {
    coder = {
      source = "coder/coder"
    }
    docker = {
      source  = "kreuzwerker/docker"
      version = "3.0.2"
    }
  }
}

variable "one" {
  default = "alice"
  type = string
}

variable "two" {
  default = "bob"
  type = string
}

variable "three" {
  default = "charlie"
  type = string
}

variable "four" {
  default = "jack"
  type = string
}


data "coder_parameter" "variable_values" {
  name        = "variable_values"
  description = "Just to show the variable values"
  type        = "string"
  default     = var.one


  option {
    name  = "one"
    value = var.one
  }

  option {
    name  = "two"
    value = var.two
  }

  option {
    name  = "three"
    value = var.three
  }

  option {
    name  = "four"
    value = var.four
  }
}
