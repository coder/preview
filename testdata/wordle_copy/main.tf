terraform {
  required_providers {
    coder = {
      source = "coder/coder"
    }
  }
}

locals {
  correct  = "lasso"
  regex    = "^[a-zA-Z]{5}$"
  errorMsg = "You must enter a 5 letter word."

  words = [
    {
      key          = "first"
      display_name = "First word"
      description  = "Enter a 5 letter word"
    },
    {
      key          = "second"
      display_name = "Second word"
      description  = "Enter the second 5 letter word"
    },
    {
      key          = "third"
      display_name = "Third word"
      description  = "Enter the third 5 letter word"
    },
    {
      key          = "fourth"
      display_name = "Fourth word"
      description  = "Enter the fourth 5 letter word"
    },
    {
      key          = "fifth"
      display_name = "Fifth word"
      description  = "Enter the fifth 5 letter word"
    },
    {
      key          = "sixth"
      display_name = "Sixth word"
      description  = "Enter the sixth 5 letter word"
    },
  ]

  valid = { for k, v in {
    first  = data.coder_parameter.words.value
    second = try(data.coder_parameter.second[0].value, "")
  } : k => length(trim(v)) == 5 }
}

data "coder_parameter" "words" {
  count = length(local.words)

  name         = local.words[count.index].key
  display_name = local.words[count.index].display_name
  description  = local.words[count.index].description
  type         = "string"
  order        = count.index + 1

  validation {
    regex = local.regex
    error = local.errorMsg
  }
}

module "words" {
  count  = length(local.words)
  source = "./checker"

  input   = data.coder_parameter.words[count.index].value
  correct = local.correct
}

output "results" {
  value = [for m in module.words : m.result]
}