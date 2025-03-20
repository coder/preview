// the correct word the player is trying to guess
variable "correct" {
  type = string
}

// The previous word
variable "previous" {
  type = string
}

locals {
  // unmatchedLetters are letters that are not exact matches from the
  // previous input.
  unmatchedLetters = [
    for i in range(0, length(var.correct)) : (
    substr(var.previous, i, 1)
    ) if substr(var.previous, i, 1) != substr(var.correct, i, 1)
  ]

  // remainingLetters are letters in the correct word that still exist to be
  // guessed.
  remainingLetters = [
    for i in range(0, length(var.correct)) : (
    substr(var.correct, i, 1)
    ) if substr(var.previous, i, 1) != substr(var.correct, i, 1)
  ]

  // letterExists are misplaced letters that exist in the correct word.
  letterExists = [
    for l in local.unmatchedLetters : (
    l
    ) if contains(local.remainingLetters, l)
  ]

  matching = join("", [
    for i in range(0, length(var.correct)) : (
      substr(var.previous, i, 1) == substr(var.correct, i, 1) ?
      upper(substr(var.correct, i, 1)) :
        contains(local.letterExists, substr(var.previous, i, 1)) ?
        lower(substr(var.previous, i, 1)) :
        "_"
    )
  ])
}

output "matching" {
  value = local.matching
}

output "debug" {
  value = {
    "correct" = var.correct
    "previous" = var.previous
    "unmatchedLetters" = local.unmatchedLetters
    "remainingLetters" = local.remainingLetters
    "letterExists" = local.letterExists
    "matching" = local.matching
  }
}