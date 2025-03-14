terraform {
  required_providers {
    coder = {
      source  = "coder/coder"
    }
  }
}

data "coder_parameter" "one" {
    count = 1
    name  = "one"
    type  = "number"
    default = 1
}

data "coder_parameter" "two" {
    count = data.coder_parameter.one[0].value
    name  = "two"
    type  = "string"
    default = "two"
}

data "coder_parameter" "three" {
    count = 1
    name  = "three"
    type  = "number"
    default = data.coder_parameter.one[0].value
}