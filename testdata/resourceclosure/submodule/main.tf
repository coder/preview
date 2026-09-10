// Submodules are never pruned; only root resources are. This module receives a
// root resource's value through an input and also owns a resource of its own.
variable "image_name" {
  type = string
}

resource "docker_image" "internal" {
  name = "sub-${var.image_name}"
}

output "image_name" {
  value = var.image_name
}

output "internal_name" {
  value = docker_image.internal.name
}
