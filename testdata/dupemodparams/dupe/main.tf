terraform {
  required_providers {
    coder = {
      version = "v2.4.0-pre0"
      source = "coder/coder"
    }
  }
}

module "dupe" {
  source = "./nesteddupe"
}

data "coder_parameter" "dupe" {
  name        = "dupe_question"
  description = "A question that will be duplicated"
  type        = "string"
  default     = "dupe"
}