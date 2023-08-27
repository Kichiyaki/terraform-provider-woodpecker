# Create a user
resource "woodpecker_user" "test" {
  login = "test"
  email = "test@localhost"
}
