terraform {
  required_providers {
    coder = {
      source = "coder/coder"
      version = "2.4.0-pre0"
    }
  }
}

data "coder_workspace_tags" "custom_workspace_tags" {
  tags = {
    "string" = "foo"
    "number" = 42
    "bool"   = true
    "list"   = ["a", "b", "c"]
    "map"    = {
      "key1" = "value1"
      "key2" = "value2"
    }
    "complex" = {
      "nested_list" = [1, 2, 3]
      "nested"  = {
        "key" = "value"
      }
    }
    "null"  = null
  }
}

