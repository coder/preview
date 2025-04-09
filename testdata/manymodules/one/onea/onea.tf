terraform {
  required_providers {
    coder = {
      version = "v2.4.0-pre0"
      source = "coder/coder"
    }
  }
}

locals {
  foo = "one-a"
}

data "coder_parameter" "one-a-question" {
  name       = "one_a_question"
  display_name = "One A Question"
  description = "From module 1, sub A"
  type        = "string"
  default     = local.foo

  option {
    name  = "Default"
    value = local.foo
  }

  option {
    name  = "Packer"
    value = jsondecode(data.http.packer.response_body).current_version
  }

  option {
    name  = "NullResource"
    value = data.null_data_source.values.outputs["foo"]
  }
}

output "export" {
  value = local.foo
}

output "packer" {
  value = jsondecode(data.http.packer.response_body).current_version
}

data "http" "packer" {
  url = "https://checkpoint-api.hashicorp.com/v1/check/packer"

  # Optional request headers
  request_headers = {
    Accept = "application/json"
    Arbitrary = data.null_data_source.values.outputs["foo"]
  }
}

data "null_data_source" "values" {
  inputs = {
    foo = "bar"
  }
}