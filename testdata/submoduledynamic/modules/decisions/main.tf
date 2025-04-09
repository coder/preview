terraform {
  required_providers {
    coder = {
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