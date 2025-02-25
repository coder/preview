terraform {
  required_providers {
    coder = {
      source = "coder/coder"
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