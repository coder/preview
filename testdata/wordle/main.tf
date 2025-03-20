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
    regex = "^(?:[A-Za-z]{5})?$"
    error = "You must enter a 5 letter word."
  }

  description = "Capital letters are an exact match, lowercase are letters that are out of place."
  alphabet = split("", "abcdefghijklmnopqrstuvwxyz")
  remove_letters =  setunion(
    try(toset(module.check_one.unmatching), []),
    try(toset(module.check_two.unmatching), []),
    try(toset(module.check_three.unmatching), []),
    try(toset(module.check_four.unmatching), []),
    try(toset(module.check_five.unmatching), []),
    try(toset(module.check_six.unmatching), []),
  )

  remaining = setsubtract(
    toset(local.alphabet),
    local.remove_letters
  )
}

output "unmatched" {
  value = toset(module.check_one.unmatching)
}

data "coder_parameter" "letter_bank" {
  name = "letter_bank"
  display_name = "Letter bank"
  description = "Remaining available letters."
  type = "string"
  order = 9
  default = join("", local.remaining)
  form_type = "input"
  # count = 0
}

data "coder_parameter" "one" {
  name = "one"
  display_name = "Take a guess what the 5 letter word might be!"
  description = "Additional guesses will appear once you input a valid 5 letter word."
  type = "string"
  order = 11
  default = ""

  form_type_metadata = jsonencode({
    disabled = length(data.coder_parameter.one.value) == 5
  })

  validation {
    regex = local.validation.regex
    error = local.validation.error
  }
}

module "check_one" {
  source = "./checker"
  correct = local.correct
  previous = data.coder_parameter.one.value
}

data "coder_parameter" "two" {
  # count = length(data.coder_parameter.one.value) == 5 ? 1 : 0
  count = 1
  name = "two"
  display_name = module.check_one.matching
  description = local.description
  type = "string"
  order = 12
  default = ""

  form_type_metadata = jsonencode({
    disabled = length(data.coder_parameter.two.value) == 5
  })

  validation {
    regex = local.validation.regex
    error = local.validation.error
  }
}

module "check_two" {
  source = "./checker"
  correct = local.correct
  previous = data.coder_parameter.two[0].value
}

output "debug" {
  value = {
    "two": length(try(data.coder_parameter.two[0].value, ""))
    "two_d": module.check_two.debug
  }
}

data "coder_parameter" "three" {
  # count = length(try(data.coder_parameter.two[0].value, "")) == 5 ? 1 : 0
  count = 1
  name = "three"
  display_name = module.check_two.matching
  description = local.description
  type = "string"
  order = 13
  default = ""

  form_type_metadata = jsonencode({
    disabled = length(data.coder_parameter.three.value) == 5
  })

  validation {
    regex = local.validation.regex
    error = local.validation.error
  }
}

module "check_three" {
  source = "./checker"
  correct = local.correct
  previous = data.coder_parameter.three[0].value
}

data "coder_parameter" "four" {
  # count = length(try(data.coder_parameter.three[0].value, "")) == 5 ? 1 : 0
  count = 1
  name = "four"
  display_name = module.check_three.matching
  description = local.description
  type = "string"
  order = 14
  default = ""

  form_type_metadata = jsonencode({
    disabled = length(data.coder_parameter.four.value) == 5
  })

  validation {
    regex = local.validation.regex
    error = local.validation.error
  }
}

module "check_four" {
  source = "./checker"
  correct = local.correct
  previous = data.coder_parameter.four[0].value
}

data "coder_parameter" "five" {
  # count = length(try(data.coder_parameter.four[0].value, "")) == 5 ? 1 : 0
  count = 1
  name = "five"
  display_name = module.check_four.matching
  description = local.description
  type = "string"
  order = 15
  default = ""

  form_type_metadata = jsonencode({
    disabled = length(data.coder_parameter.five.value) == 5
  })

  validation {
    regex = local.validation.regex
    error = local.validation.error
  }
}

module "check_five" {
  source = "./checker"
  correct = local.correct
  previous = data.coder_parameter.five[0].value
}

data "coder_parameter" "six" {
  # count = length(try(data.coder_parameter.five[0].value, "")) == 5 ? 1 : 0
  count = 1
  name = "six"
  display_name = module.check_five.matching
  description = local.description
  type = "string"
  order = 16
  default = ""

  form_type_metadata = jsonencode({
    disabled = length(data.coder_parameter.six.value) == 5
  })

  validation {
    regex = local.validation.regex
    error = local.validation.error
  }
}

module "check_six" {
  source = "./checker"
  correct = local.correct
  previous = data.coder_parameter.six[0].value
}

#
# module "word_one" {
#   source = "./checker"
#   input = data.coder_parameter.first.value
#   correct = local.correct
# }
#
# data "coder_parameter" "first" {
#   name         = "first"
#   display_name = "First word"
#   description  = "Enter a 5 letter word"
#   type         = "string"
#   order        = 1
#
#   validation {
#     regex = local.validation.regex
#     error = local.validation.error
#   }
# }
#
# // ---
# module "word_two" {
#   source = "./checker"
#   input = data.coder_parameter.second[0].value
#   correct = local.correct
# }
#
# data "coder_parameter" "second" {
#   count = 1
#   # count        = length(data.coder_parameter.first.value) == 5 ? 1 : 0
#
#   name         = "second"
#   display_name = "Second word"
#   description  = "Previous word matches: ${module.word_one.result}"
#   type         = "string"
#   order        = 2
#
#   validation {
#     regex = local.validation.regex
#     error = local.validation.error
#   }
# }
#
# // ---
# module "word_three" {
#   source = "./checker"
#   input = data.coder_parameter.third[0].value
#   correct = local.correct
# }
#
# data "coder_parameter" "third" {
#   count = 1
#   # count        = try(length(data.coder_parameter.second[0].value) == 5, false) ? 1 : 0
#
#   name         = "third"
#   display_name = "Third word"
#   description  = "Previous word matches: ${module.word_two.result}"
#   type         = "string"
#   order        = 3
#
#   validation {
#     regex = local.validation.regex
#     error = local.validation.error
#   }
# }
#
# // ---
# module "word_four" {
#   source = "./checker"
#   input = data.coder_parameter.fourth[0].value
#   correct = local.correct
# }
#
# data "coder_parameter" "fourth" {
#   count = 1
#   name         = "fourth"
#   display_name = "Fourth word"
#   description  = "Previous word matches: ${module.word_three.result}"
#   type         = "string"
#   order        = 4
#
#   validation {
#     regex = local.validation.regex
#     error = local.validation.error
#   }
# }
#
# // ---
# module "word_five" {
#   source = "./checker"
#   input = data.coder_parameter.fifth[0].value
#   correct = local.correct
# }
#
# data "coder_parameter" "fifth" {
#   count = 1
#   name         = "fifth"
#   display_name = "Fifth word"
#   description  = "Previous word matches: ${module.word_four.result}"
#   type         = "string"
#   order        = 5
#
#   validation {
#     regex = local.validation.regex
#     error = local.validation.error
#   }
# }
#
# // ---
# # module "word_six" {
# #   source = "./checker"
# #   input = data.coder_parameter.fifth.value
# #   correct = local.correct
# # }
#
# data "coder_parameter" "six" {
#   count = 1
#   name         = "sixth"
#   display_name = "Sixth word"
#   description  = "Previous word matches: ${module.word_five.result}"
#   type         = "string"
#   order        = 6
#
#   validation {
#     regex = local.validation.regex
#     error = local.validation.error
#   }
# }