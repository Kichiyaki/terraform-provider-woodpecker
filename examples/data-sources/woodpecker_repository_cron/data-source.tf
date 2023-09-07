resource "woodpecker_repository" "test_repo" {
  full_name = "Kichiyaki/test-repo"
}

data "woodpecker_repository_secret" "test_secret" {
  repository_id = woodpecker_repository.test_repo.id
  id            = 1
}
