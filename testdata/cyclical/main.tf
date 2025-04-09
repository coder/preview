terraform {
  required_providers {
    coder = {
      version = "v2.4.0-pre0"
      source = "coder/coder"
    }
  }
}

data "coder_parameter" "alpha" {
  name        = "alpha"
  description = "Alpha parameter"
  type        = "string"
  default     = data.coder_parameter.beta.value
  order       = 1
}

data "coder_parameter" "beta" {
  name        = "beta"
  description = "Beta parameter"
  type        = "string"
  default     = data.coder_parameter.alpha.value
  order       = 2
}
