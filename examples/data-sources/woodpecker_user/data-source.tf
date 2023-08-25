# Retrieve information about the currently authenticated user.
data "woodpecker_user" "current" {
  login = ""
}

# Retrieve information about a user.
data "woodpecker_user" "user" {
  login = "user"
}
