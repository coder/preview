terraform {
  required_providers {
    coder = {
      source = "coder/coder"
    }
  }
}

# The feature flags below are module outputs, so they are unknown during the
# first expansion passes and only become known after this module is evaluated.
module "gate" {
  source = "./modules/gate"
}

# `count` resolves late, so the module block is cloned after its own blocks
# captured their ModuleBlock() pointer.
module "ai" {
  source = "./modules/ai"
  count  = module.gate.enabled ? 1 : 0
}

# Same, through `for_each`.
module "regional" {
  source   = "./modules/regional"
  for_each = toset(module.gate.regions)
  region   = each.value
}

module "plain" {
  source = "./modules/plain"
}
