resource "woodpecker_repository" "test_repo" {
  full_name  = "Kichiyaki/test-repo"
  is_trusted = true
  visibility = "public"
}

resource "woodpecker_repository_cron" "test" {
  repository_id = woodpecker_repository.test_repo.id
  name          = "test"
  schedule      = "@daily"
}
