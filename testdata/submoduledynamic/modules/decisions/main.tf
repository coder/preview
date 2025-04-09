terraform {
  required_providers {
    coder = {
      version = "v2.4.0-pre0"
      source = "coder/coder"
    }
  }
}

output "zero" {
  value = 0
}

output "no" {
  value = false
}

output "yes" {
  value = true
}