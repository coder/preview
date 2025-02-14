terraform {
  required_providers {
    coder = {
      source  = "coder/coder"
    }
  }
}

locals {
  foo = "one"
}

data "coder_parameter" "onequestion" {
  name        = "One Question"
  description = "From module 1"
  type        = "string"
  default     = local.foo

  option {
      name  = "Default"
      value = local.foo
  }

  option {
    name  = "Primary Sub A"
    value = module.onea.export
  }

  option {
    name  = "Terraform"
    value = jsondecode(data.http.terraform.response_body).current_version
  }
}

module "onea" {
  source = "./onea"
}

output "export" {
  value = local.foo
}

output "export-a" {
  value = module.onea.export
}

output "terraform" {
  value = jsondecode(data.http.terraform.response_body).current_version
}

data "http" "terraform" {
  url = "https://checkpoint-api.hashicorp.com/v1/check/terraform"

  # Optional request headers
  request_headers = {
    Accept = "application/json"
  }
}