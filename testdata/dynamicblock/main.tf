terraform {
  required_providers {
    coder = {
      source = "coder/coder"
      version = "2.4.0-pre0"
    }
  }
}

variable "regions" {
  type    = set(string)
  default = ["us", "eu", "au"]
}

data "coder_parameter" "indexed" {
  count = 2
  name = "indexed_${count.index}"
  display_name = "Indexed Param ${count.index}"
  type = "string"
  form_type = "dropdown"
  default = "Hello_0"

  dynamic "option" {
    for_each = range(2)

    content {
      name  = "Hello_${option.value}"
      value = "Hello_${option.value}"
    }
  }
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
