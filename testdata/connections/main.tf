terraform {
  required_providers {
    coder = {
      source = "coder/coder"
    }
  }
}

locals {
  word_bank = [
    // Outspoken -- Yellow
    "direct", "frank", "loud", "vocal",
    // Bodies of water -- Green
    "bay", "channel", "sound", "strait",
    // Kinds of cords -- Blue
    "bungee", "extension", "spinal", "umbilical",
    // Things in bottles -- Purple
    "genie", "lighting", "message", "ship"
  ]

  used_words = setunion(
    [],
    jsondecode(data.coder_parameter.yellow.value),
    jsondecode(data.coder_parameter.green.value),
  )

  available_words = setsubtract(toset(local.word_bank), toset(local.used_words))
}

data "coder_parameter" "yellow" {
  name = "yellow"
  display_name = "Row"
  type = "list(string)"
  form_type = "multi-select"
  # default = "[]"

  dynamic "option" {
    # for_each = tolist(setsubtract(toset(local.word_bank), toset(local.used_words)))
    for_each = local.word_bank
    content {
      name = option.value
      value = option.value
    }
  }
}

data "coder_parameter" "green" {
  name = "green"
  display_name = "Row"
  type = "list(string)"
  form_type = "multi-select"
  # default = "[]"

  dynamic "option" {
    # for_each = tolist(setsubtract(toset(local.word_bank), toset(local.used_words)))
    for_each = local.word_bank
    content {
      name = option.value
      value = option.value
    }
  }
}

output "remaining" {
  value = local.available_words
}

output "used" {
  value = local.used_words
}

output "yellow" {
  value = data.coder_parameter.yellow.value
}

