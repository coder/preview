/*

go run ../../cmd/preview/main.go \
-v "\"yellow=[\"\"bay\"\",\"\"sound\"\",\"\"strait\"\", \"\"channel\"\"]\"" \
-v "\"green=[\"\"bungee\"\",\"\"extension\"\",\"\"spinal\"\", \"\"umbilical\"\"]\"" \
-v "\"blue=[\"\"direct\"\",\"\"loud\"\",\"\"vocal\"\", \"\"frank\"\"]\"" \
-v "\"purple=[\"\"ship\"\",\"\"genie\"\",\"\"lighting\"\", \"\"message\"\"]\""
*/

terraform {
  required_providers {
    coder = {
      source = "coder/coder"
      version = "v2.4.0-pre0"
    }
  }
}

locals {
  solutions = tomap ({
    // Outspoken -- Yellow
    "Outspoken": ["direct", "frank", "loud", "vocal"],
    // Bodies of water -- Green
    "Bodies of water": ["bay", "channel", "sound", "strait"],
    // Kinds of cords -- Blue
    "Kinds of cords": ["bungee", "extension", "spinal", "umbilical"],
    // Things in bottles -- Purple
    "Things in a bottle": ["genie", "lighting", "message", "ship"],
  })
  # solution_list = [for _, words in local.solutions : words]
  word_bank = flatten([for _, words in local.solutions : words])


  used_words = setunion(
    [],
    jsondecode(data.coder_parameter.rows["yellow"].value),
    jsondecode(data.coder_parameter.rows["green"].value),
    jsondecode(data.coder_parameter.rows["blue"].value),
    jsondecode(data.coder_parameter.rows["purple"].value),
  )

  available_words = setsubtract(toset(local.word_bank), toset(local.used_words))

  colors = toset(["yellow", "green", "blue", "purple"])

  solved = length([for color in local.colors : module.checker[color].solved if module.checker[color].solved]) == 4
}



module "checker" {
  for_each = local.colors
  source = "./checker"
  solutions = local.solutions
  guess = jsondecode(coalesce(data.coder_parameter.rows[each.value].value, "[]"))
}

data "coder_parameter" display {
  name = "display"
  display_name = local.solved ? "Congrats, you won! You may now hit the switch!" : join(", ", local.available_words)
  description = local.solved ? "Hitting the switch enables workspace creation." : "Remaining words are above, you cannot use this switch until you solve the puzzle!"
  type = "bool"
  form_type = "switch"
  default = local.solved ? false : true
  # default = local.solved ? "" : "Keep guessing!"

  styling = jsonencode({
    disabled = !local.solved
  })
}

output "solved" {
  value = local.solved
}


data "coder_parameter" "rows" {
  for_each = local.colors
  name = each.value
  display_name = module.checker[each.value].title
  description = module.checker[each.value].description
  # name = "rows"
  type = "list(string)"
  form_type = "multi-select"
  styling = jsonencode({
    disabled = module.checker[each.value].solved
  })
  default = "[]"
  order = 11

  dynamic "option" {
    # for_each = toset(local.word_bank)
    // Must include the options that are selected, otherwise they are not in
    // the option set.
    for_each = toset(concat(tolist(local.available_words), jsondecode(data.coder_parameter.rows[each.value].value)))
    content {
      name = option.value
      value = option.value
    }
  }

  # validation {
  #   error = "Hey! ${length(data.coder_parameter.rows[each.value].value)}"
  #   invalid = true
  # }
}



