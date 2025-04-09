terraform {
  required_providers {
    coder = {
      version = "v2.4.0-pre0"
      source = "coder/coder"
    }
  }
}

variable "regions" {
  type    = set(string)
  default = ["us", "eu", "au"]
}

data "coder_parameter" "region" {
  name        = "Region"
  description = "Which region would you like to deploy to?"
  type        = "string"
  default     = tolist(var.regions)[0]
  
  dynamic "option" {
    for_each = var.regions
    content {
      name  = option.value
      value = option.value
    }
  }
}

data "coder_workspace_tags" "custom_workspace_tags" {
  tags = {
    "zone" = data.coder_parameter.region.value
  }
}
