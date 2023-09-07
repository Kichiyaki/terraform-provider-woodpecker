resource "woodpecker_repository" "test_repo" {
  full_name  = "Kichiyaki/test-repo"
  is_trusted = true
  visibility = "public"
}

resource "woodpecker_repository_secret" "test" {
  repository_id = woodpecker_repository.test_repo.id
  address       = "docker.io"
  username      = "test"
  password      = "test"
}
