terraform {
  required_providers {
    coder = {
      source = "coder/coder"
      version = "2.4.0-pre0"
    }
  }
}

module "dupe" {
  source = "./dupe"
}

data "coder_parameter" "dupe" {
  name        = "dupe_question"
  description = "A question that will be duplicated"
  type        = "string"
  default     = "dupe"
}