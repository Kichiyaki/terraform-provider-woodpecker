package internal_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestRepositoryDataSource(t *testing.T) {
	t.Parallel()

	repo := createRepo(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "woodpecker_repository" "test_repo" {
	full_name = "%s"
}

data "woodpecker_repository" "test_repo" {
	full_name = woodpecker_repository.test_repo.full_name
}
`, repo.FullName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.woodpecker_repository.test_repo", "id"),
					resource.TestCheckResourceAttr(
						"data.woodpecker_repository.test_repo",
						"forge_remote_id",
						strconv.FormatInt(repo.ID, 10),
					),
					resource.TestCheckResourceAttr("data.woodpecker_repository.test_repo", "name", repo.Name),
					resource.TestCheckResourceAttr("data.woodpecker_repository.test_repo", "owner", repo.Owner.UserName),
					resource.TestCheckResourceAttr("data.woodpecker_repository.test_repo", "full_name", repo.FullName),
					resource.TestCheckResourceAttr("data.woodpecker_repository.test_repo", "url", repo.HTMLURL),
					resource.TestCheckResourceAttr("data.woodpecker_repository.test_repo", "clone_url", repo.CloneURL),
					resource.TestCheckResourceAttr("data.woodpecker_repository.test_repo", "default_branch", repo.DefaultBranch),
					resource.TestCheckResourceAttr("data.woodpecker_repository.test_repo", "scm", "git"),
					resource.TestCheckResourceAttrSet("data.woodpecker_repository.test_repo", "timeout"),
					resource.TestCheckResourceAttr("data.woodpecker_repository.test_repo", "visibility", "public"),
					resource.TestCheckResourceAttr(
						"data.woodpecker_repository.test_repo",
						"is_private",
						strconv.FormatBool(repo.Private),
					),
					resource.TestCheckResourceAttr("data.woodpecker_repository.test_repo", "is_trusted", "false"),
					resource.TestCheckResourceAttr("data.woodpecker_repository.test_repo", "is_gated", "false"),
					resource.TestCheckResourceAttr("data.woodpecker_repository.test_repo", "allow_pull_requests", "true"),
					resource.TestCheckResourceAttr("data.woodpecker_repository.test_repo", "config_file", ""),
					resource.TestCheckResourceAttr("data.woodpecker_repository.test_repo", "netrc_only_trusted", "true"),
				),
			},
		},
	})
}
