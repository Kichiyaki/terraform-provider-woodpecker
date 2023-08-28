# Create a secret
resource "woodpecker_secret" "test" {
  name   = "test"
  value  = "test"
  events = ["cron", "deployment"]
}
