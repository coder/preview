terraform {
  required_providers {
    coder = {
      source = "coder/coder"
    }
  }
}

module "one" {
  source = "./one"
}

module "two" {
  source = "./two"
}

locals {
    foo = "main"
}

data "coder_parameter" "mainquestion" {
  name        = "Two Question"
  description = "From module 2"
  type        = "string"
  default     = local.foo
}