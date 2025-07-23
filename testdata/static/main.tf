// Base case for workspace tags + parameters.
terraform {
  required_providers {
    coder = {
      source = "coder/coder"
      version = "2.4.0-pre0"
    }
    docker = {
      source  = "kreuzwerker/docker"
      version = "3.0.2"
    }
  }
}

variable "string" {
  default = "Hello, world!"
  nullable = true
  sensitive =  true
  description = "test"
}

variable "number" {
  default = 7
}

variable "boolean" {
  default = true
}

variable "coerce_string" {
  default = 5 // This will be coerced to a string
  type = string
}

variable "complex" {
  type = object({
    list   = list(string)
    name = string
    age  = number
  })
  default = {
    list = []
    name = "John Doe"
    age  = 30
  }
}


data "coder_workspace_tags" "custom_workspace_tags" {
  tags = {
    "zone"        = "developers"
    "null"        = null
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
