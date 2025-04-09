terraform {
  required_providers {
    coder = {
      version = "v2.4.0-pre0"
      source = "coder/coder"
    }
  }
}

locals {
  string_opts = [
    {
      name        = "Alpha"
      value       = "alpha-value"
      description = "This is option one"
      icon        = "/emojis/0031-fe0f-20e3.png"
    },
    {
      name        = "Bravo"
      value       = "bravo-value"
      description = "This is option two"
      icon        = "/emojis/0032-fe0f-20e3.png"
    },
    {
      name        = "Charlie"
      value       = "charlie-value"
      description = "This is option three"
      icon        = "/emojis/0033-fe0f-20e3.png"
    }
  ]
}


data "coder_parameter" "single_select" {
  name         = "single_select"
  display_name = "How do you want to format the options of the next parameter?"
  description  = "The next parameter supports a single value."
  type         = "string"
  form_type    = "dropdown"
  order        = 10
  default      = "radio"

  option {
    name        = "Radio Selector"
    value       = "radio"
    description = "Radio selections."
    icon        = "/emojis/0031-fe0f-20e3.png"
  }

  option {
    name        = "Dropdown Selector"
    value       = "dropdown"
    description = "Dropdown selections."
    icon        = "/emojis/0031-fe0f-20e3.png"
  }

  option {
    name        = "Raw Input"
    value       = "input"
    description = "Input whatever you want."
    icon        = "/emojis/0031-fe0f-20e3.png"
  }

  option {
    name        = "Multiline input"
    value       = "textarea"
    description = "A larger text area."
    icon        = "/emojis/0031-fe0f-20e3.png"
  }
}

data "coder_parameter" "single" {
  name         = "single"
  display_name = "Selecting a single value from a list of options."
  description  = "Change the formatting of this parameter with the param above."
  type         = "string"
  order        = 11
  icon         = "/emojis/0031-fe0f-20e3.png"
  default      = local.string_opts[0].value
  form_type    = data.coder_parameter.single_select.value

  dynamic "option" {
    for_each = data.coder_parameter.single_select.value == "input" || data.coder_parameter.single_select.value == "textarea" ? [] : local.string_opts
    content {
      name        = option.value.name
      value       = option.value.value
      description = option.value.description
      icon        = option.value.icon
    }
  }
}

data "coder_parameter" "number_format" {
  name         = "number_format"
  display_name = "How do you want to format the options of the next parameter?"
  description  = "The next parameter supports numerical values."
  type         = "string"
  form_type    = "dropdown"
  order        = 20
  default      = "input"

  option {
    name        = "Slider"
    value       = "slider"
    description = "Slider."
    icon        = "/emojis/0031-fe0f-20e3.png"
  }

  option {
    name        = "Raw input"
    value       = "input"
    description = "Type in a number."
    icon        = "/emojis/0031-fe0f-20e3.png"
  }
}

data "coder_parameter" "number" {
  name         = "number"
  display_name = "What is your favorite number?"
  description  = "Change the formatting of this parameter with the param above."
  type         = "number"
  order        = 21
  icon         = "/emojis/0031-fe0f-20e3.png"
  default      = 7
  form_type    = data.coder_parameter.number_format.value

  validation {
    min   = 0
    max   = 100
    error = "Value {value} is not between {min} and {max}"
  }
}

data "coder_parameter" "boolean_format" {
  name         = "boolean_format"
  display_name = "How do you want to format the options of the next parameter?"
  description  = "The next parameter supports boolean values."
  type         = "string"
  form_type    = "dropdown"
  order        = 30
  default      = "radio"

  option {
    name        = "Radio"
    value       = "radio"
    description = "Radio."
    icon        = "/emojis/0031-fe0f-20e3.png"
  }

  option {
    name        = "Switch"
    value       = "switch"
    description = "Switch."
    icon        = "/emojis/0031-fe0f-20e3.png"
  }

  option {
    name        = "Checkbox"
    value       = "checkbox"
    description = "Checkbox."
    icon        = "/emojis/0031-fe0f-20e3.png"
  }
}

locals {
  boolean_opts = [
    {
      name        = "Yes"
      value       = true
      description = "Yes, I agree to the terms."
      icon        = "/emojis/0031-fe0f-20e3.png"
    },
    {
      name        = "No"
      value       = false
      description = "No, I do not agree to the terms."
      icon        = "/emojis/0032-fe0f-20e3.png"
    }
  ]
}

data "coder_parameter" "boolean" {
  name         = "boolean"
  display_name = "Do you agree with me?"
  description  = "Selecting true is the best choice."
  type         = "bool"
  order        = 31
  icon         = "/emojis/0031-fe0f-20e3.png"
  default      = true
  form_type    = data.coder_parameter.boolean_format.value

  dynamic "option" {
    for_each = data.coder_parameter.boolean_format.value == "radio" ? local.boolean_opts : []
    content {
      name        = option.value.name
      value       = option.value.value
      description = option.value.description
      icon        = option.value.icon
    }
  }
}

data "coder_parameter" "list_format" {
  name         = "list_format"
  display_name = "How do you want to format the options of the next parameter?"
  description  = "The next parameter supports lists of values."
  type         = "string"
  form_type    = "dropdown"
  order        = 40
  default      = "multi-select"


  option {
    name        = "Multi-Select"
    value       = "multi-select"
    description = "Select multiple."
    icon        = "/emojis/0031-fe0f-20e3.png"
  }

  option {
    name        = "Radio"
    value       = "radio"
    description = "Radio."
    icon        = "/emojis/0031-fe0f-20e3.png"
  }

  option {
    name        = "Tag Select"
    value       = "tag-select"
    description = "Tag select."
    icon        = "/emojis/0031-fe0f-20e3.png"
  }
}

locals {
  radio_list_opts = [
    {
      name        = "None"
      value       = jsonencode([])
      description = "Red related colors."
      icon        = "/emojis/0031-fe0f-20e3.png"
    },
    {
      name        = "Reds"
      value       = jsonencode(["red"])
      description = "Red related colors."
      icon        = "/emojis/0031-fe0f-20e3.png"
    },
    {
      name = "Blue & Green"
      value = jsonencode(["blue", "green"])
      description = "Blue and green related colors."
      icon = "/emojis/0031-fe0f-20e3.png"
    }
  ]
  list_color_opts = [
    {
      name        = "Red"
      value       = "red"
      description = "The color of blood."
      icon        = "/emojis/0031-fe0f-20e3.png"
    },
    {
      name        = "Orange"
      value       = "orange"
      description = "The color of oranges."
      icon        = "/emojis/0032-fe0f-20e3.png"
    },
    {
      name        = "Yellow"
      value       = "yellow"
      description = "The color of the sun."
      icon        = "/emojis/0033-fe0f-20e3.png"
    },
    {
      name        = "Green"
      value       = "green"
      description = "The color of grass."
      icon        = "/emojis/0034-fe0f-20e3.png"
    },
    {
      name        = "Blue"
      value       = "blue"
      description = "The color of the sky."
      icon        = "/emojis/0035-fe0f-20e3.png"
    },
    {
      name        = "Purple"
      value       = "purple"
      description = "The color of royalty."
      icon        = "/emojis/0036-fe0f-20e3.png"
    }
  ]
}

data "coder_parameter" "list" {
  name         = "list"
  display_name = "What colors are the best?"
  description  = "Select a few if you want."
  type         = "list(string)"
  order        = 41
  icon         = "/emojis/0031-fe0f-20e3.png"
  default      = jsonencode(["blue", "green"])
  form_type    = data.coder_parameter.list_format.value

  dynamic "option" {
    for_each = data.coder_parameter.list_format.value == "radio" ? local.radio_list_opts : (
      data.coder_parameter.list_format.value == "tag-select" ? [] : local.list_color_opts
    )
    content {
      name        = option.value.name
      value       = option.value.value
      description = option.value.description
      icon        = option.value.icon
    }
  }
}

data "coder_parameter" "like_it" {
  name         = "like_it"
  display_name = "Did you like this demo?"
  description  = "Please check!"
  type         = "bool"
  form_type    = "checkbox"
  order        = 50
  default      = false
}

data "coder_parameter" "satisfaction" {
  count = data.coder_parameter.like_it.value ? 1 : 0
  name         = "satisfaction"
  display_name = "Please rate your satisfaction."
  description  = ""
  type         = "number"
  form_type    = "slider"
  order        = 51
  default      = 85

  validation {
    min   = 0
    max   = 100
    error = "Value {value} is not between {min} and {max}"
  }
}
