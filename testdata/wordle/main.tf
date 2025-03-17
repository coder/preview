terraform {
  required_providers {
    coder = {
      source = "coder/coder"
    }
  }
}

locals {
  correct = "lasso" // March 17, 2025
  validation = {
    regex = "^[a-zA-Z]{5}$"
    error = "You must enter a 5 letter word."
  }
}

module "word_one" {
  source = "./checker"
  input = data.coder_parameter.first.value
  correct = local.correct
}

data "coder_parameter" "first" {
  name         = "first"
  display_name = "First word"
  description  = "Enter a 5 letter word"
  type         = "string"
  order        = 1

  validation {
    regex = local.validation.regex
    error = local.validation.error
  }
}

// ---
module "word_two" {
  source = "./checker"
  input = data.coder_parameter.second[0].value
  correct = local.correct
}

data "coder_parameter" "second" {
  count = 1
  # count        = length(data.coder_parameter.first.value) == 5 ? 1 : 0

  name         = "second"
  display_name = "Second word"
  description  = "Previous word matches: ${module.word_one.result}"
  type         = "string"
  order        = 2

  validation {
    regex = local.validation.regex
    error = local.validation.error
  }
}

// ---
module "word_three" {
  source = "./checker"
  input = data.coder_parameter.third[0].value
  correct = local.correct
}

data "coder_parameter" "third" {
  count = 1
  # count        = try(length(data.coder_parameter.second[0].value) == 5, false) ? 1 : 0

  name         = "third"
  display_name = "Third word"
  description  = "Previous word matches: ${module.word_two.result}"
  type         = "string"
  order        = 3

  validation {
    regex = local.validation.regex
    error = local.validation.error
  }
}

// ---
module "word_four" {
  source = "./checker"
  input = data.coder_parameter.fourth[0].value
  correct = local.correct
}

data "coder_parameter" "fourth" {
  count = 1
  name         = "fourth"
  display_name = "Fourth word"
  description  = "Previous word matches: ${module.word_three.result}"
  type         = "string"
  order        = 4

  validation {
    regex = local.validation.regex
    error = local.validation.error
  }
}

// ---
module "word_five" {
  source = "./checker"
  input = data.coder_parameter.fifth[0].value
  correct = local.correct
}

data "coder_parameter" "fifth" {
  count = 1
  name         = "fifth"
  display_name = "Fifth word"
  description  = "Previous word matches: ${module.word_four.result}"
  type         = "string"
  order        = 5

  validation {
    regex = local.validation.regex
    error = local.validation.error
  }
}

// ---
# module "word_six" {
#   source = "./checker"
#   input = data.coder_parameter.fifth.value
#   correct = local.correct
# }

data "coder_parameter" "six" {
  count = 1
  name         = "sixth"
  display_name = "Sixth word"
  description  = "Previous word matches: ${module.word_five.result}"
  type         = "string"
  order        = 6

  validation {
    regex = local.validation.regex
    error = local.validation.error
  }
}