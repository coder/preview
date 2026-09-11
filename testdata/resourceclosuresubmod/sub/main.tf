terraform {
  required_providers {
    coder = {
      source  = "coder/coder"
      version = "2.4.0-pre0"
    }
  }
}

variable "image_name" {
  type = string
}

data "coder_parameter" "flavor" {
  name    = "flavor"
  type    = "string"
  default = var.image_name
}
