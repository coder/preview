terraform {
  required_providers {
    coder = {
      source = "coder/coder"
      version = "2.8.0"
    }
  }
}

data "coder_parameter" "use_custom_image" {
  name    = "use_custom_image"
  type    = "bool"
  default = "false"
}

data "coder_parameter" "custom_image_url" {
  count   = data.coder_parameter.use_custom_image.value ? 1 : 0
  name    = "custom_image_url"
  type    = "string"
  # No default - required when shown
}

data "coder_workspace_preset" "valid_preset" {
  name = "valid_preset"
  parameters = {
    "use_custom_image" = "true"
    "custom_image_url" = "docker.io/codercom/test:latest"
  }
  prebuilds {
    instances = 1
  }
}