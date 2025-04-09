terraform {
  required_providers {
    coder = {
      source = "coder/coder"
      version = "v2.4.0-pre0"
    }
  }
}

data "coder_parameter" "dupe" {
  name        = "Dupe Question"
  description = "A question that will be duplicated"
  type        = "string"
  default     = "dupe"
}

data "coder_parameter" "dupe" {
  name        = "Dupe Question"
  description = "A question that will be duplicated"
  type        = "string"
  default     = "dupe"
}