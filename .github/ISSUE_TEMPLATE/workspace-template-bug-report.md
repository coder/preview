---
name: Workspace Template Bug Report
about: Workspace template yielded incorrect parameters.
title: "[BUG] Workspace template behavior"
labels: ''
assignees: Emyrk

---

**Describe the bug**
A clear and concise description of what the bug is. Was the parameter value/description/options/etc incorrect? Did you encounter an error you did not expect? Did you expect to encounter an error, and did not?

**Expected behavior**
A clear and concise description of what you expected to happen. What was the parameter supposed to look like?

**Offending Template**
Provide the most minimal workspace template that reproduces the bug. Try to remove any non-coder terraform blocks. Only `data "coder_parameter"` and `data "coder_workspace_tags"`  with any supporting or referenced blocks are required.

If the template is a single `main.tf`, please include the `main.tf` in the collapsible section below. If there are multiple files, either attach the files, or create a public github repository with the directory structure. Try to avoid attaching zips or tarballs.


<details>

<summary>Template `main.tf`</summary>

```terraform
# Replace this with your `main.tf`
terraform {
  required_providers {
    coder = {
      source = "coder/coder"
      version = "2.3.0"
    }
  }
}

data "coder_parameter" "region" {
  name        = "region"
  description = "Which region would you like to deploy to?"
  type        = "string"
  default     = "us"
  order       = 1

  option {
    name  = "Europe"
    value = "eu"
  }
  option {
    name  = "United States"
    value = "us"
  }
}
```

</details>

**Screenshots**
If applicable, add screenshots to help explain your problem.

**Tooling (please complete the following information):**
 - Terraform Version: [e.g. v1.11.2]
 - Coderd Version [e.g. chrome, v2.20.2]
 - Coder Provider Version [e.g. 2.3.0, if not in the `main.tf`]

**Additional context**
Add any other context about the problem here.
