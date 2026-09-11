variable "image_name" {
  type = string
}

output "name" {
  value = "inner-${var.image_name}"
}
