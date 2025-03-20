terraform {
  required_providers {
    coder = {
      source = "coder/coder"
    }
  }
}

locals {
  string_opts = [
    {
      name        = "one"
      value       = "one"
      description = "This is option one"
      icon        = "/emojis/0031-fe0f-20e3.png"
    },
    {
      name        = "two"
      value       = "two"
      description = "This is option two"
      icon        = "/emojis/0032-fe0f-20e3.png"
    },
    {
      name        = "three"
      value       = "three"
      description = "This is option three"
      icon        = "/emojis/0033-fe0f-20e3.png"
    }
  ]
}

data "coder_parameter" "string_opts_default" {
  // should be 'radio'
  name         = "string_opts"
  display_name = "String"
  description  = "String with options"
  type         = "string"
  order        = 1
  icon         = "/emojis/0031-fe0f-20e3.png"

  dynamic "option" {
    for_each = local.string_opts
    content {
      name        = option.value.name
      value       = option.value.value
      description = option.value.description
      icon        = option.value.icon
    }
  }
}

data "coder_parameter" "string_opts_dropdown" {
  name         = "string_opts_default"
  display_name = "String"
  description  = "String with options and default"
  type         = "string"
  form_type    = "dropdown"
  order        = 2
  icon         = "/emojis/0031-fe0f-20e3.png"

  dynamic "option" {
    for_each = local.string_opts
    content {
      name        = option.value.name
      value       = option.value.value
      description = option.value.description
      icon        = option.value.icon
    }
  }
}

data "coder_parameter" "string_without_opts" {
  // should be 'input'
  name         = "string_without_opts"
  display_name = "String"
  description  = "String without options"
  type         = "string"
  order        = 3
  icon         = "/emojis/0031-fe0f-20e3.png"
}

data "coder_parameter" "textarea_without_opts" {
  name         = "textarea"
  display_name = "String"
  description  = "Textarea"
  type         = "string"
  form_type    = "textarea"
  order        = 4
  icon         = "/emojis/0031-fe0f-20e3.png"
}

data "coder_parameter" "bool_with_opts" {
  // should be 'radio'
  name         = "bool_with_opts"
  display_name = "Bool"
  description  = "Bool with options"
  type         = "bool"
  order        = 5
  icon         = "/emojis/0031-fe0f-20e3.png"
  default      = false

  option {
    name        = "Yes"
    value       = true
    description = "Yes, I agree to the terms."
  }

  option {
    name        = "No"
    value       = false
    description = "No, I do not agree to the terms."
  }
}

data "coder_parameter" "bool_without_opts" {
  // should be 'checkbox'
  name         = "bool_without_opts"
  display_name = "Bool"
  description  = "Bool without options"
  type         = "bool"
  order        = 6
  icon         = "/emojis/0031-fe0f-20e3.png"
  default      = false
}

data "coder_parameter" "bool_without_opts_switch" {
  name         = "bool_without_opts_switch"
  display_name = "Bool"
  description  = "Bool without options, but it is a switch"
  type         = "bool"
  form_type    = "switch"
  order        = 7
  icon         = "/emojis/0031-fe0f-20e3.png"
  default      = false
}

data "coder_parameter" "list_string_options" {
  // should be radio
  name         = "list_string_options"
  display_name = "List(String)"
  description  = "list(string) with options"
  type         = "list(string)"
  order        = 8
  icon         = "/emojis/0031-fe0f-20e3.png"
  default      = jsonencode(["purple", "blue", "green", "red", "orange"])

  option {
    name        = "All"
    description = "All the colors"
    value       = jsonencode(["purple", "blue", "green", "red", "orange"])
  }

  option {
    name        = "Bluish Colors"
    description = "Colors that are kinda blue"
    value       = jsonencode(["purple", "blue"])
  }

  option {
    name        = "Redish Colors"
    description = "Colors that are kinda red"
    value       = jsonencode(["red", "orange"])
  }
}

data "coder_parameter" "list_string_without_options" {
  // should be tag-select
  name         = "list_string_without_options"
  display_name = "List(String)"
  description  = "list(string) with options"
  type         = "list(string)"
  order        = 9
  icon         = "/emojis/0031-fe0f-20e3.png"
  default      = jsonencode(["purple", "blue", "green", "red", "orange"])
  // You could send jsonencode(["airplane", "car", "school"])
}

data "coder_parameter" "list_string_multi_select_options" {
  // should be multi-select
  name         = "list_string_multi_select_options"
  display_name = "List(String)"
  description  = "list(string) with options"
  type         = "list(string)"
  form_type    = "multi-select"
  order        = 10
  icon         = "/emojis/0031-fe0f-20e3.png"
  default      = jsonencode(["blue", "green", "red"])

  option {
    name        = "Blue"
    value       = "blue"
    description = "Like the sky."
  }

  option {
    name        = "Red"
    value       = "red"
    description = "Like a rose."
  }

  option {
    name        = "Green"
    value       = "green"
    description = "Like the grass."
  }

  option {
    name        = "Purple"
    value       = "purple"
    description = "Like a grape."
  }

  option {
    name        = "Orange"
    value       = "orange"
    description = "Like the fruit."
  }
}



