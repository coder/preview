variable "image_name" {
  type = string
}

module "inner" {
  source     = "./inner"
  image_name = "outer-${var.image_name}"
}

output "inner_name" {
  value = module.inner.name
}
