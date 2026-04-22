data "coder_secret" "gh" {
  env          = "GITHUB_TOKEN"
  help_message = "Add a GitHub PAT"
}

data "coder_secret" "aws" {
  file         = "~/.aws/credentials"
  help_message = "Add AWS creds"
}
