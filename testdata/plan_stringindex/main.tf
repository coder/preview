terraform {
  required_providers {
    coder = {
      source = "coder/coder"
    }
    docker = {
      source  = "kreuzwerker/docker"
      version = "~> 3.0.0"
    }
    envbuilder = {
      source = "coder/envbuilder"
    }
  }
}

variable "jetbrains_ides" {
  type        = list(string)
  description = "The list of IDE product codes."
  default     = ["IU", "PS", "WS", "PY", "CL", "GO", "RM", "RD", "RR"]
  validation {
    condition = (
    alltrue([
      for code in var.jetbrains_ides : contains(["IU", "PS", "WS", "PY", "CL", "GO", "RM", "RD", "RR"], code)
    ])
    )
    error_message = "The jetbrains_ides must be a list of valid product codes. Valid product codes are ${join(",", ["IU", "PS", "WS", "PY", "CL", "GO", "RM", "RD", "RR"])}."
  }
  # check if the list is empty
  validation {
    condition     = length(var.jetbrains_ides) > 0
    error_message = "The jetbrains_ides must not be empty."
  }
  # check if the list contains duplicates
  validation {
    condition     = length(var.jetbrains_ides) == length(toset(var.jetbrains_ides))
    error_message = "The jetbrains_ides must not contain duplicates."
  }
}

variable "releases_base_link" {
  type        = string
  description = ""
  default     = "https://data.services.jetbrains.com"
  validation {
    condition     = can(regex("^https?://.+$", var.releases_base_link))
    error_message = "The releases_base_link must be a valid HTTP/S address."
  }
}

variable "channel" {
  type        = string
  description = "JetBrains IDE release channel. Valid values are release and eap."
  default     = "release"
  validation {
    condition     = can(regex("^(release|eap)$", var.channel))
    error_message = "The channel must be either release or eap."
  }
}


data "http" "jetbrains_ide_versions" {
  for_each = toset(var.jetbrains_ides)
  url      = "${var.releases_base_link}/products/releases?code=${each.key}&latest=true&type=${var.channel}"
}

variable "download_base_link" {
  type        = string
  description = ""
  default     = "https://download.jetbrains.com"
  validation {
    condition     = can(regex("^https?://.+$", var.download_base_link))
    error_message = "The download_base_link must be a valid HTTP/S address."
  }
}

variable "arch" {
  type        = string
  description = "The target architecture of the workspace"
  default     = "amd64"
  validation {
    condition     = contains(["amd64", "arm64"], var.arch)
    error_message = "Architecture must be either 'amd64' or 'arm64'."
  }
}

variable "jetbrains_ide_versions" {
  type = map(object({
    build_number = string
    version      = string
  }))
  description = "The set of versions for each jetbrains IDE"
  default = {
    "IU" = {
      build_number = "243.21565.193"
      version      = "2024.3"
    }
    "PS" = {
      build_number = "243.21565.202"
      version      = "2024.3"
    }
    "WS" = {
      build_number = "243.21565.180"
      version      = "2024.3"
    }
    "PY" = {
      build_number = "243.21565.199"
      version      = "2024.3"
    }
    "CL" = {
      build_number = "243.21565.238"
      version      = "2024.1"
    }
    "GO" = {
      build_number = "243.21565.208"
      version      = "2024.3"
    }
    "RM" = {
      build_number = "243.21565.197"
      version      = "2024.3"
    }
    "RD" = {
      build_number = "243.21565.191"
      version      = "2024.3"
    }
    "RR" = {
      build_number = "243.22562.230"
      version      = "2024.3"
    }
  }
  validation {
    condition = (
    alltrue([
      for code in keys(var.jetbrains_ide_versions) : contains(["IU", "PS", "WS", "PY", "CL", "GO", "RM", "RD", "RR"], code)
    ])
    )
    error_message = "The jetbrains_ide_versions must contain a map of valid product codes. Valid product codes are ${join(",", ["IU", "PS", "WS", "PY", "CL", "GO", "RM", "RD", "RR"])}."
  }
}

locals {
  # AMD64 versions of the images just use the version string, while ARM64
  # versions append "-aarch64". Eg:
  #
  # https://download.jetbrains.com/idea/ideaIU-2025.1.tar.gz
  # https://download.jetbrains.com/idea/ideaIU-2025.1.tar.gz
  #
  # We rewrite the data map above dynamically based on the user's architecture parameter.
  #
  effective_jetbrains_ide_versions = {
    for k, v in var.jetbrains_ide_versions : k => {
      build_number = v.build_number
      version      = var.arch == "arm64" ? "${v.version}-aarch64" : v.version
    }
  }

  # When downloading the latest IDE, the download link in the JSON is either:
  #
  # linux.download_link
  # linuxARM64.download_link
  #
  download_key = var.arch == "arm64" ? "linuxARM64" : "linux"

  jetbrains_ides = {
    "GO" = {
      icon          = "/icon/goland.svg",
      name          = "GoLand",
      identifier    = "GO",
      version       = local.effective_jetbrains_ide_versions["GO"].version
    },
    "WS" = {
      icon          = "/icon/webstorm.svg",
      name          = "WebStorm",
      identifier    = "WS",
      version       = local.effective_jetbrains_ide_versions["WS"].version
    },
    "IU" = {
      icon          = "/icon/intellij.svg",
      name          = "IntelliJ IDEA Ultimate",
      identifier    = "IU",
      version       = local.effective_jetbrains_ide_versions["IU"].version
    },
    "PY" = {
      icon          = "/icon/pycharm.svg",
      name          = "PyCharm Professional",
      identifier    = "PY",
      version       = local.effective_jetbrains_ide_versions["PY"].version
    },
    "CL" = {
      icon          = "/icon/clion.svg",
      name          = "CLion",
      identifier    = "CL",
      version       = local.effective_jetbrains_ide_versions["CL"].version
    },
    "PS" = {
      icon          = "/icon/phpstorm.svg",
      name          = "PhpStorm",
      identifier    = "PS",
      version       = local.effective_jetbrains_ide_versions["PS"].version
    },
    "RM" = {
      icon          = "/icon/rubymine.svg",
      name          = "RubyMine",
      identifier    = "RM",
      version       = local.effective_jetbrains_ide_versions["RM"].version
    },
    "RD" = {
      icon          = "/icon/rider.svg",
      name          = "Rider",
      identifier    = "RD",
      version       = local.effective_jetbrains_ide_versions["RD"].version
    },
    "RR" = {
      icon          = "/icon/rustrover.svg",
      name          = "RustRover",
      identifier    = "RR",
      version       = local.effective_jetbrains_ide_versions["RR"].version
    }
  }
}
data "coder_parameter" "jetbrains_ide" {
  type         = "string"
  name         = "jetbrains_ide"
  display_name = "JetBrains IDE"
  icon         = "/icon/gateway.svg"
  mutable      = true
  default      = var.jetbrains_ides[0]

  dynamic "option" {
    for_each = var.jetbrains_ides
    content {
      icon  = local.jetbrains_ides[option.value].icon
      name  = "${local.jetbrains_ides[option.value].name} ${local.jetbrains_ides[option.value].version}"
      value = option.value
    }
  }
}
