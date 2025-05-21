terraform {
  required_providers {
    coder = {
      source = "coder/coder"
      version = "2.4.0-pre0"
    }
  }
}


data "coder_parameter" "file" {
  name        = "file"
  description = "Attempt to read some files."
  type        = "string"
  order       = 1

  option {
    name  = "Local"
    value = file("./hello.txt")
  }

  option {
    name  = "Outer"
    value = file("../README.md")
  }
}
