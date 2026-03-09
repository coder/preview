terraform {
  required_providers {
    coder = {
      source = "coder/coder"
      version = "2.5.3"
    }
  }
}

locals {
  vscode_name = "VSCode"
  ides = [
    {
      name  = local.vscode_name,
      value = "vscode",
    },
  ]
}

data "coder_parameter" "ide_picker" {
  name      = "ide_picker"
  type      = "list(string)"
  form_type = "multi-select"
  default   = jsonencode(["vscode"])
  dynamic "option" {
    for_each = local.ides
    content {
      name  = option.value.name
      value = option.value.value
    }
  }
}
