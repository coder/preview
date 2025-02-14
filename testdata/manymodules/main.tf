terraform {
  required_providers {
    coder = {
      source = "coder/coder"
    }
  }
}

module "one" {
  source = "./one"
}

module "two" {
  source = "./two"
}

locals {
    foo = "main"
}

data "coder_parameter" "mainquestion" {
  name        = "Main Question"
  description = "From module 2"
  type        = "string"
  default     = local.foo

  option {
    name  = "Default"
    value = local.foo
  }

  option {
    name  = "Primary"
    value = module.one.export
  }

  option {
    name  = "Second"
    value = module.two.export
  }

  option {
    name  = "Terraform"
    value = module.one.terraform
  }

  option {
    name  = "Consul"
    value = module.two.consul
  }

  option {
    name  = "Packer"
    value = module.one.export-a
  }
}