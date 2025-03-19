module "zero" {
  source = "./zero"
}

module "param" {
  count = module.zero.zero
  source = "./param"
}