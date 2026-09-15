// Observes whether resources outside the parameter/preset/tag closure were
// evaluated. Nothing a parameter reads references the resource, so the default
// closure skips it and the output that reads it is unknown. With
// OptionFullEvaluation the resource is evaluated and the output resolves.
terraform {
  required_providers {
    coder = {
      source  = "coder/coder"
      version = "2.4.0-pre0"
    }
    docker = {
      source  = "kreuzwerker/docker"
      version = "3.0.2"
    }
  }
}

data "coder_parameter" "flavor" {
  name    = "flavor"
  type    = "string"
  default = "large"
}

resource "docker_image" "unreferenced" {
  name = "outside-closure"
}

output "unreferenced_name" {
  value = docker_image.unreferenced.name
}
