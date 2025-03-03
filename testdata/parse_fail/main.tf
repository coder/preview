locals {
  example = (true
}

data "coder_parameter" "example" {
  name        = "Example"
  type        = "string"
  default     = "foo"
}