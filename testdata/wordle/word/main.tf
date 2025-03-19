terraform {
  required_providers {
    coder = {
      source = "coder/coder"
    }
  }
}

locals {
  names = ["first", "second", "third", "fourth", "fifth", "sixth"]
  capitalized = "${upper(substr(local.names[var.index], 0, 1))}${substr(local.names[var.index], 1, length(local.names[var.index]) - 1)}"
}

// what word is this?
variable "index" {
  type = number
}

// the correct word the player is trying to guess
variable "correct" {
  type = string
}

// The previous word
variable "previous" {
  type = string
}

// optional for debugging
variable "default" {
  type = string
  default = ""
}

output "value" {
  value = data.coder_parameter.word[0].value
}

locals {
  matching = join("", [
    for i in range(0, length(var.correct)) : (
      substr(var.previous, i, 1) == substr(var.correct, i, 1) ? "#" : "_"
    )
  ])

  // unmatchedLetters are letters that are not exact matches from the
  // previous input.
  unmatchedLetters = [
    for i in range(0, length(var.correct)) : (
      substr(var.previous, i, 1)
    ) if substr(var.previous, i, 1) != substr(var.correct, i, 1)
  ]

  // remainingLetters are letters in the correct word that still exist to be
  // guessed.
  remainingLetters = [
    for i in range(0, length(var.correct)) : (
    substr(var.correct, i, 1)
    ) if substr(var.previous, i, 1) != substr(var.correct, i, 1)
  ]

  // letterExists are misplaced letters that exist in the correct word.
  letterExists = [
    for l in local.unmatchedLetters : (
      l
    ) if contains(local.remainingLetters, l)
  ]
}

data "coder_parameter" "word" {
  count        = length(var.previous) == 5 ? 1 : 0
  name         = local.names[var.index]
  display_name = "--> ${local.matching} <-- with ${join("", local.letterExists)}"
  description = local.matching
  type         = "string"
  order        = var.index + 10
  default      = var.default

  validation {
    regex = "^[a-zA-Z]{5}$"
    error = "You must enter a 5 letter word."
  }
}


