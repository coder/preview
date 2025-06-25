// Base case for workspace tags + parameters.
terraform {
  required_providers {
    coder = {
      source = "coder/coder"
      version = "2.4.0-pre0"
    }
  }
}

data "coder_parameter" "region" {
  name        = "region"
  description = "Which region would you like to deploy to?"
  type        = "string"
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