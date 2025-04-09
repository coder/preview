terraform {
  required_providers {
    coder = {
      source = "coder/coder"
      version = "v2.4.0-pre0"
    }
  }
}

module "sub" {
  source = "./submodule"
}

data "coder_workspace_tags" "test" {
  tags = {
    "test" = tostring(module.sub.static == "static")
  }
}

data "coder_parameter" "region" {
  count       = module.sub.static == "static" ? 1 : 0
  name        = "Region"
  description = "Which region would you like to deploy to?"
  type        = "string"
  default     = upper(module.sub.static)
}
