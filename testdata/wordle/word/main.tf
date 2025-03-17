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

// The pattern for showing the previous correct letters.
variable "pattern" {
  type = string
}

// optional for debugging
variable "default" {
  type = string
  default = ""
}

data "coder_parameter" "word" {
  name         = local.names[var.index]
  display_name = "${local.capitalized} word ${module.checker.valid}"
  description  = var.pattern
  type         = "string"
  order        = var.index + 10
  default      = var.default

  validation {
    regex = "^[a-zA-Z]{5}$"
    error = "You must enter a 5 letter word."
  }
}

module "checker" {
  source = "../checker"
  input = data.coder_parameter.word.value
  correct = var.correct
}

output "result" {
  value = module.checker.result
}

output "valid" {
  value = module.checker.valid
  # value = length(data.coder_parameter.word.value) == 5
}