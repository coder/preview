variable "input" {
  type = string
}

variable "correct" {
  type = string
}

output "result" {
  value = join("", [
    for i in range(0, length(var.correct)) : (
      substr(var.input, i, 1) == substr(var.correct, i, 1) ? "#" : "_"
    )
  ])
}