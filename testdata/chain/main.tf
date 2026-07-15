terraform {
  required_providers {
    coder = {
      source = "coder/coder"
      version = "2.5.3"
    }
  }
}

locals {
  ides = [
    "VS Code",
    "JetBrains IntelliJ",
    "GoLand",
    "WebStorm",
    "PyCharm",
    "Databricks",
    "Jupyter Notebook",
  ]

  is_ml_repo = data.coder_parameter.git_repo == "coder/mlkit"

  selected = try(jsondecode(data.coder_parameter.ide_selector[0].value), [])
}


data "coder_parameter" "git_repo" {
  name = "git_repo"
  display_name = "Git repo"
  description = "Select a git repo to work on."
  order = 1
  mutable = true
  type = "string"
  form_type = "dropdown"

  option {
    # A Go-heavy repository
    name = "coder/coder"
    value = "coder/coder"
  }

  option {
    # A python-heavy repository
    name = "coder/mlkit"
    value = "coder/mlkit"
  }
}

data "coder_parameter" "ide_selector" {
  count = try(data.coder_parameter.git_repo.value, "") != "" ? 1 : 0
  name = "ide_selector"
  description  = "Choose any IDEs for your workspace."
  mutable      = true
  display_name = "Select mutliple IDEs"
  order = 1
  default = "[]"

  # Allows users to select multiple IDEs from the list.
  form_type = "multi-select"
  type      = "list(string)"


  dynamic "option" {
    for_each = local.ides
    content {
      name  = option.value
      value = option.value
    }
  }
}


data "coder_parameter" "cpu_cores" {
  # Only show this parameter if the previous box is selected.
  count = length(local.selected) > 0 ? 1 : 0

  name         = "cpu_cores"
  display_name = "CPU Cores"
  type         = "number"
  form_type    = "slider"
  default      = local.is_ml_repo ? 12 : 6
  order        = 2
  validation {
    min = 1
    max = local.is_ml_repo ? 16 : 8
  }
}

output "selected" {
  value = local.selected
}

output "static" {
  value = "foo"
}