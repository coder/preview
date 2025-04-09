terraform {
  required_providers {
    coder = {
      source = "coder/coder"
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