resource "woodpecker_repository" "test_repo" {
  full_name  = "%s"
  is_trusted = true
  visibility = "public"
}
