data "woodpecker_org" "test_org" {
  name = "test"
}

resource "woodpecker_org_secret" "test" {
  org_id = data.woodpecker_org.test_org.id
  name   = "test"
  value  = "test"
  events = ["cron", "deployment"]
}
