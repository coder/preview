terraform {
  required_providers {
    coder = {
      source = "coder/coder"
      version = "2.4.0-pre0"
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