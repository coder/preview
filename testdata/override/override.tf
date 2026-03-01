# Override region's default again (sequential: us -> eu -> ap).
data "coder_parameter" "region" {
  default = "ap"

  option {
    name  = "AP"
    value = "ap"
  }
}

# Override preset.
data "coder_workspace_preset" "dev" {
  name = "dev-override"
  parameters = {
    region = "ap"
  }
}
