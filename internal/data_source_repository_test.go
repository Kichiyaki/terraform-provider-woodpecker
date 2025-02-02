package internal_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/Kichiyaki/terraform-provider-woodpecker/internal/woodpecker"
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
					resource.TestCheckResourceAttrSet("data.woodpecker_repository.test_repo", "forge_id"),
					resource.TestCheckResourceAttr(
						"data.woodpecker_repository.test_repo",
						"forge_remote_id",
						strconv.FormatInt(repo.ID, 10),
					),
					resource.TestCheckResourceAttr("data.woodpecker_repository.test_repo", "forge_url", repo.HTMLURL),
					resource.TestCheckResourceAttr("data.woodpecker_repository.test_repo", "clone_url", repo.CloneURL),
					resource.TestCheckResourceAttr("data.woodpecker_repository.test_repo", "owner", repo.Owner.UserName),
					resource.TestCheckResourceAttr("data.woodpecker_repository.test_repo", "name", repo.Name),
					resource.TestCheckResourceAttr("data.woodpecker_repository.test_repo", "full_name", repo.FullName),
					resource.TestCheckResourceAttrSet("data.woodpecker_repository.test_repo", "avatar_url"),
					resource.TestCheckResourceAttr("data.woodpecker_repository.test_repo", "default_branch", repo.DefaultBranch),
					resource.TestCheckResourceAttrSet("data.woodpecker_repository.test_repo", "timeout"),
					resource.TestCheckResourceAttr(
						"data.woodpecker_repository.test_repo",
						"visibility",
						woodpecker.VisibilityModePublic.String(),
					),
					resource.TestCheckResourceAttr(
						"data.woodpecker_repository.test_repo",
						"is_private",
						strconv.FormatBool(repo.Private),
					),
					resource.TestCheckResourceAttr("data.woodpecker_repository.test_repo", "trusted.network", "false"),
					resource.TestCheckResourceAttr("data.woodpecker_repository.test_repo", "trusted.security", "false"),
					resource.TestCheckResourceAttr("data.woodpecker_repository.test_repo", "trusted.volumes", "false"),
					resource.TestCheckResourceAttr(
						"data.woodpecker_repository.test_repo",
						"require_approval",
						woodpecker.ApprovalModeForks.String(),
					),
					resource.TestCheckResourceAttr("data.woodpecker_repository.test_repo", "is_active", "true"),
					resource.TestCheckResourceAttr("data.woodpecker_repository.test_repo", "allow_pull_requests", "true"),
					resource.TestCheckResourceAttr("data.woodpecker_repository.test_repo", "allow_deployments", "false"),
					resource.TestCheckResourceAttr("data.woodpecker_repository.test_repo", "config_file", ""),
					resource.TestCheckTypeSetElemAttr(
						"data.woodpecker_repository.test_repo",
						"cancel_previous_pipeline_events.*",
						woodpecker.EventPush,
					),
					resource.TestCheckTypeSetElemAttr(
						"data.woodpecker_repository.test_repo",
						"cancel_previous_pipeline_events.*",
						woodpecker.EventPull,
					),
					resource.TestCheckNoResourceAttr("data.woodpecker_repository.test_repo", "netrc_trusted_plugins"),
				),
			},
		},
	})
}
