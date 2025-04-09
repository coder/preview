terraform {
  required_providers {
    coder = {
      source = "coder/coder"
      version = "2.4.0-pre0"
    }
  }
}

locals {
  foo = "two"
}

data "coder_parameter" "twoquestion" {
  name = "two_question"
  display_name = "Two Question"
  description = "From module 2"
  type        = "string"
  default     = local.foo

  option {
    name  = "Default"
    value = local.foo
  }

  option {
    name  = "Consul"
    value = jsondecode(data.http.consul.response_body).current_version
  }
}

output "export" {
  value = local.foo
}

output "consul" {
  value = jsondecode(data.http.consul.response_body).current_version
}

data "http" "consul" {
  url = "https://checkpoint-api.hashicorp.com/v1/check/consul"

  # Optional request headers
  request_headers = {
    Accept = "application/json"
  }
}