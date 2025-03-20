
variable "solutions" {
  type = map(list(string))
}

variable "guess" {
  type = list(string)
}

locals {
# [for connection, solution in local.solutions : connection if (length(setintersection(solution, jsondecode(data.coder_parameter.rows["yellow"].value))) == 4)]
  diff = [for connection, solution in var.solutions : {
    connection = connection
    distance = 4 - length(setintersection(solution, var.guess))
  }]

  solved = [for diff in local.diff : diff.connection if diff.distance == 0]
  one_away = [for diff in local.diff : diff.connection if diff.distance == 1]
  description = length(local.one_away) == 1 ? "One away..." : (
      length(local.solved) == 1 ? "Solved!" : (
      "Select 4 words that share a common connection."
    )
  )
}

output "out" {
  value = local.one_away
}

output "title" {
  value = length(local.solved) == 1 ? "${local.solved[0]}" : "??"
}

output "description" {
  value = local.description
}

output "solved" {
  value = length(local.solved) == 1 ? true : false
}