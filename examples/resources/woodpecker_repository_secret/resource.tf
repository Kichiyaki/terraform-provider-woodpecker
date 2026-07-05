resource "woodpecker_repository" "test_repo" {
  full_name  = "Kichiyaki/test-repo"
  is_trusted = true
  visibility = "public"
}

resource "woodpecker_repository_secret" "test" {
  repository_id = woodpecker_repository.test_repo.id
  name          = "test"
  value         = "test"
  events        = ["cron", "deployment"]
}

# Supply the value as a write-only attribute so it's never persisted in state
# (requires Terraform 1.11+ or OpenTofu 1.11+). Bump value_wo_version whenever
# the value changes to push the new value to Woodpecker.
variable "ci_token" {
  type      = string
  sensitive = true
}

resource "woodpecker_repository_secret" "write_only" {
  repository_id    = woodpecker_repository.test_repo.id
  name             = "write_only"
  value_wo         = var.ci_token
  value_wo_version = 1
  events           = ["push"]
}
