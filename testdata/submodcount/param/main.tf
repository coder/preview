terraform {
  required_providers {
    coder = {
      source = "coder/coder"
    }
  }
}

data "coder_parameter" "example" {
  name         = "example"
  display_name = "Example"
  description  = "An example parameter"
  type         = "string"
  order        = 1
  default      = "example"
}