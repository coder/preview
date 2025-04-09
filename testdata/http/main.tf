terraform {
  required_providers {
    coder = {
      version = "v2.4.0-pre0"
      source = "coder/coder"
    }
  }
}

data "http" "example" {
  url = "https://checkpoint-api.hashicorp.com/v1/check/terraform"

  # Optional request headers
  request_headers = {
    Accept = "application/json"
  }
}

data "coder_workspace_tags" "custom_workspace_tags" {
  tags = {
    "tfversion" = jsondecode(data.http.example.response_body)["current_version"]
  }
}