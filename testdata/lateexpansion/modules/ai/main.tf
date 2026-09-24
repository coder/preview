terraform {
  required_providers {
    coder = {
      source = "coder/coder"
    }
  }
}

data "coder_parameter" "ai_model" {
  name    = "ai_model"
  type    = "string"
  default = "sonnet"
}
