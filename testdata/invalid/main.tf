terraform {
  required_providers {
    coder = {
      source = "coder/coder"
      version = "2.4.0-pre0"
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
