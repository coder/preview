// Base case for workspace tags + parameters.
terraform {
  required_providers {
    coder = {
      version = "v2.4.0-pre0"
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
    "zone"        = "developers"
  }
}

data "coder_parameter" "region" {
  name        = "region"
  description = "Which region would you like to deploy to?"
  type        = "string"
  default     = "us"
  order       = 1

  option {
    name  = "Europe"
    value = "eu"
  }
  option {
    name  = "United States"
    value = "us"
  }
}

data "coder_parameter" "numerical" {
  name        = "numerical"
  description = "Numerical parameter"
  type        = "number"
  default     = 5
  order       = 2
}
