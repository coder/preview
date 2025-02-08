terraform {
  required_providers {
    coder = {
      source  = "coder/coder"
    }
  }
}

locals {
  foo = "one"
}

data "coder_parameter" "onequestion" {
  name        = "One Question"
  description = "From module 1"
  type        = "string"
  default     = local.foo
}