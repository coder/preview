terraform {
  required_providers {
    coder = {
      source = "coder/coder"
    }
    docker = {
      source  = "kreuzwerker/docker"
      version = "3.0.2"
    }
  }
}

data "coder_workspace_tags" "custom_workspace_tags" {
  tags = {
    "zone"        = data.coder_parameter.region.value
  }
}

data "coder_parameter" "region" {
  name        = "Region"
  description = "Which region would you like to deploy to?"
  type        = "string"
  default     = "us"

  option {
    name  = "Europe"
    value = "eu"
    description = ""
  }
  option {
    name  = "United States"
    value = "us"
  }
}