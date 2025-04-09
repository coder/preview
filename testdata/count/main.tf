terraform {
  required_providers {
    coder = {
      source = "coder/coder"
    }
    null = {
      source  = "hashicorp/null"
      version = "3.2.2"
    }
  }
}

data "null_data_source" "exists" {
  # for_each = toset(["one", "two", "three"])
  count = 3
  inputs = {
    foo = "Index ${count.index}"
  }
}

data "coder_parameter" "ref" {
  name = "ref"
  type = "string"
  default = data.null_data_source.exists[2].inputs.foo
  count = 1

  option {
    name  = "exists 1"
    value = data.null_data_source.exists[0].inputs.foo
  }
  option {
    name  = "exists 2"
    value = data.null_data_source.exists[1].inputs.foo
  }
  option {
    name  = "exists 3"
    value = data.null_data_source.exists[2].inputs.foo
  }
}

data "coder_parameter" "ref_count" {
  name = "ref_count"
  type = "string"
  default = data.coder_parameter.ref[0].default
  count = 1

  option {
    name  = "Only"
    value = data.coder_parameter.ref[0].default
  }
}
