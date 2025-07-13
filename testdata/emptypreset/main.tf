terraform {
  required_providers {
    coder = {
      source = "coder/coder"
      version = "2.8.0"
    }
  }
}

data "coder_workspace_preset" "test" {
}