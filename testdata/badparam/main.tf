terraform {
  required_providers {
    coder = {
      source = "coder/coder"
      version = "v2.4.0-pre0"
    }
  }
}


data "coder_parameter" "region" {
}