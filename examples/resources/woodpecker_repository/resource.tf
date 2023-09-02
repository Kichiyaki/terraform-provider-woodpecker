resource "woodpecker_repository" "test_repo" {
  full_name  = "Kichiyaki/test-repo"
  is_trusted = true
  visibility = "public"
}
