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
    "zone" = 5
    10     = "hello"
  }
}

