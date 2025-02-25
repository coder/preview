terraform {
  required_providers {
    coder = {
      source = "coder/coder"
    }
  }
}

module "dupe" {
  source = "./nesteddupe"
}

data "coder_parameter" "dupe" {
  name        = "Dupe Question"
  description = "A question that will be duplicated"
  type        = "string"
  default     = "dupe"
}