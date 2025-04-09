terraform {
  required_providers {
    coder = {
      version = "v2.4.0-pre0"
      source = "coder/coder"
    }
  }
}

data "coder_parameter" "invalid" {
  name        = "invalid"
  type        = "string"
  default     = "random"
  order       = 1

  validation {
    invalid = true
    error   = "This is an invalid parameter"
  }
}
