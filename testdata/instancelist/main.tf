terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    coder = {
      source = "coder/coder"
    }
  }
}

data "coder_parameter" "user_location" {
  name        = "Home"
  description = "Where do you live?"
  type        = "string"
  default     = "us"

  option {
    name  = "United States"
    value = "us"
  }
  option {
    name  = "Europe"
    value = "eu"
  }
}

variable "regions" {
  type = list(string)
  default = ["us-east-1", "us-west-2", "us-west-1", "eu-south-1", "eu-west-1"]
}

locals {
  limit_instance_types = slice(tolist(data.aws_ec2_instance_type_offerings.example.instance_types),0,20)
  allowed_regions = [for region in var.regions : region if can(regex("^${data.coder_parameter.user_location.value}-", region))]
}

data "coder_parameter" "region" {
    name        = "Region"
    description = "Which region would you like to deploy to?"
    type        = "string"
    default     = local.allowed_regions[0]

    validation {
      regex = "^${data.coder_parameter.user_location.value}-"
      error = "Region must start with the same value as your home location"
    }

    dynamic "option" {
      for_each = local.allowed_regions
      content {
        name  = option.value
        value = option.value
      }
    }
}

data "coder_parameter" "instance_type" {
    name = "instance_type"
    display_name        = "Instance Type"
    description = "Which instance type would you like to use?"
    type        = "string"
    default     = local.limit_instance_types[0]

    dynamic "option" {
        # for_each = ["us-east-1", "us-west-2", "us-west-1"]
        for_each = local.limit_instance_types
        content {
        name  = option.value
        value = option.value
        }
    }
}

provider "aws" {
  region = data.coder_parameter.region.value
}


data "aws_ec2_instance_type_offerings" "example" {
  location_type = "region"
}
