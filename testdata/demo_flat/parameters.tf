data "coder_parameter" "team" {
  name        = "team"
  display_name = "Team"
  description = "Which team are you on?"
  type        = "string"
  default     = "fullstack"
  order       = 10

  dynamic "option" {
    for_each = local.teams
    content {
      name  = option.value.display_name
      value = option.key
      description = option.value.description
      icon = option.value.icon
    }
  }

  validation {
    regex = "^frontend|backend|fullstack$"
    error = "You must select either frontend, backend, or fullstack."
  }
}

data "coder_parameter" "browser" {
  name        = "browser"
  description = "Which browser do you prefer?"
  type        = "string"
  default     = "chromium"
  order       = 12
  # count = 1
  count       = (
    data.coder_parameter.team.value == "frontend" ||
    data.coder_parameter.team.value == "fullstack"? 1 : 0
  )

  option {
    name  = "Chrome"
    value = "chrome"
  }

  option {
    name  = "Firefox"
    value = "firefox"
  }

  option {
    name  = "Safari"
    value = "safari"
  }

  option {
    name  = "Edge"
    value = "edge"
  }

  option {
    name  = "Chromium"
    value = "chromium"
  }
}


data "coder_parameter" "cpu" {
  name         = "cpu"
  display_name = "CPU"
  description  = "The number of CPU cores"
  form_type    = "slider"
  type         = "number"
  default      = "2"
  icon         = "/icon/memory.svg"
  mutable      = true
  order        = 20

  styling = jsonencode({
    "minumum" = 1
    "budget"  = 2
    "performance" = local.secutity_level == "high" ? 4 : 8
  })

  validation {
    min = 1
    // Confidential instances are more expensive, or some justification like
    // that
    // TODO: This breaks when the user is an admin
    max = local.secutity_level == "high" ? 4 : 8
    error = "CPU range must be between {min} and {max}."
  }
}

data "coder_workspace_tags" "test" {
    tags = {
      "hash": trimprefix(data.docker_registry_image.coder.sha256_digest, "sha256:")
    }
}

// Advanced admin parameter
// preview -g admin -p plan.json -v hash="52bb4d943694f2f5867a251780f85e5a68906787b4ffa3157e29b9ef510b1a97"
data "coder_parameter" "image_hash" {
  count       = local.isAdmin ? 1 : 0
  name        = "hash"
  display_name = "Image Hash"
  description = "Override the hash of the image to use. Only available to admins."
  // Value can get stale
  default     = trimprefix(data.docker_registry_image.coder.sha256_digest, "sha256:")
  order       = 100

  validation {
    regex = "^[a-f0-9A-F]{64}$"
    error = "The image hash must be a 64-character hexadecimal string."
  }
}

data "docker_registry_image" "coder" {
  name = "ghcr.io/coder/coder:latest"
}

data "coder_parameter" "region" {
  name         = "region"
  display_name = "Region"
  description  = "What region are you in?"
  form_type    = "dropdown"
  default      = local.default_region
  icon         = "/icon/memory.svg"
  mutable      = false
  order        = 1

  dynamic "option" {
    for_each = local.regions
    content {
      name  = option.value.name
      value = option.value.value
      icon  = option.value.icon
    }
  }
}
