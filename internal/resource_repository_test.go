package internal_test

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"testing"

	"github.com/Kichiyaki/terraform-provider-woodpecker/internal/woodpecker"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestRepositoryResource(t *testing.T) {
	t.Parallel()

	t.Run("OK", func(t *testing.T) {
		t.Parallel()

		repo1, repo2 := createRepo(t), createRepo(t)
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			CheckDestroy:             checkRepositoryResourceDestroy(repo1.FullName, repo2.FullName),
			Steps: []resource.TestStep{
				{ // create repo
					Config: fmt.Sprintf(`
resource "woodpecker_repository" "test_repo" {
	full_name = "%s"
	visibility = "%s"
}
`, repo1.FullName, woodpecker.VisibilityModePublic),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("woodpecker_repository.test_repo", "id"),
						resource.TestCheckResourceAttrSet("woodpecker_repository.test_repo", "forge_id"),
						resource.TestCheckResourceAttr(
							"woodpecker_repository.test_repo",
							"forge_remote_id",
							strconv.FormatInt(repo1.ID, 10),
						),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "forge_url", repo1.HTMLURL),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "clone_url", repo1.CloneURL),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "owner", repo1.Owner.UserName),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "name", repo1.Name),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "full_name", repo1.FullName),
						resource.TestCheckResourceAttrSet("woodpecker_repository.test_repo", "avatar_url"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "default_branch", repo1.DefaultBranch),
						resource.TestCheckResourceAttrSet("woodpecker_repository.test_repo", "timeout"),
						resource.TestCheckResourceAttr(
							"woodpecker_repository.test_repo",
							"visibility",
							woodpecker.VisibilityModePublic.String(),
						),
						resource.TestCheckResourceAttr(
							"woodpecker_repository.test_repo",
							"is_private",
							strconv.FormatBool(repo1.Private),
						),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "trusted.network", "false"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "trusted.security", "false"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "trusted.volumes", "false"),
						resource.TestCheckResourceAttr(
							"woodpecker_repository.test_repo",
							"require_approval",
							woodpecker.ApprovalModeForks.String(),
						),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "is_active", "true"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "allow_pull_requests", "true"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "allow_deployments", "false"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "config_file", ""),
						resource.TestCheckTypeSetElemAttr(
							"woodpecker_repository.test_repo",
							"cancel_previous_pipeline_events.*",
							woodpecker.EventPush,
						),
						resource.TestCheckTypeSetElemAttr(
							"woodpecker_repository.test_repo",
							"cancel_previous_pipeline_events.*",
							woodpecker.EventPull,
						),
						resource.TestCheckNoResourceAttr("woodpecker_repository.test_repo", "netrc_trusted_plugins"),
					),
				},
				{ // update repo
					Config: fmt.Sprintf(`
resource "woodpecker_repository" "test_repo" {
	full_name = "%s"
	visibility = "%s"
	timeout = 30
	config_file = ".woodpecker2.yaml"
	trusted = {
		network = true	
	}
	allow_deployments = true
	allow_pull_requests = false
}
`, repo1.FullName, woodpecker.VisibilityModePrivate),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("woodpecker_repository.test_repo", "id"),
						resource.TestCheckResourceAttrSet("woodpecker_repository.test_repo", "forge_id"),
						resource.TestCheckResourceAttr(
							"woodpecker_repository.test_repo",
							"forge_remote_id",
							strconv.FormatInt(repo1.ID, 10),
						),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "forge_url", repo1.HTMLURL),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "clone_url", repo1.CloneURL),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "owner", repo1.Owner.UserName),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "name", repo1.Name),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "full_name", repo1.FullName),
						resource.TestCheckResourceAttrSet("woodpecker_repository.test_repo", "avatar_url"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "default_branch", repo1.DefaultBranch),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "timeout", "30"),
						resource.TestCheckResourceAttr(
							"woodpecker_repository.test_repo",
							"visibility",
							woodpecker.VisibilityModePrivate.String(),
						),
						resource.TestCheckResourceAttr(
							"woodpecker_repository.test_repo",
							"is_private",
							strconv.FormatBool(repo1.Private),
						),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "trusted.network", "true"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "trusted.security", "false"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "trusted.volumes", "false"),
						resource.TestCheckResourceAttr(
							"woodpecker_repository.test_repo",
							"require_approval",
							woodpecker.ApprovalModeForks.String(),
						),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "is_active", "true"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "allow_pull_requests", "false"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "allow_deployments", "true"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "config_file", ".woodpecker2.yaml"),
						resource.TestCheckTypeSetElemAttr(
							"woodpecker_repository.test_repo",
							"cancel_previous_pipeline_events.*",
							woodpecker.EventPush,
						),
						resource.TestCheckTypeSetElemAttr(
							"woodpecker_repository.test_repo",
							"cancel_previous_pipeline_events.*",
							woodpecker.EventPull,
						),
						resource.TestCheckNoResourceAttr("woodpecker_repository.test_repo", "netrc_trusted_plugins"),
					),
				},
				{ // update repo
					Config: fmt.Sprintf(`
resource "woodpecker_repository" "test_repo" {
	full_name = "%s"
	trusted = {
		volumes = true	
	}
	cancel_previous_pipeline_events = ["%s"]
	netrc_trusted_plugins = ["woodpeckerci/plugin-docker-buildx"]
	approval_allowed_users = ["%s"]
}
`, repo1.FullName, woodpecker.EventTag, repo1.Owner.UserName),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("woodpecker_repository.test_repo", "id"),
						resource.TestCheckResourceAttrSet("woodpecker_repository.test_repo", "forge_id"),
						resource.TestCheckResourceAttr(
							"woodpecker_repository.test_repo",
							"forge_remote_id",
							strconv.FormatInt(repo1.ID, 10),
						),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "forge_url", repo1.HTMLURL),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "clone_url", repo1.CloneURL),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "owner", repo1.Owner.UserName),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "name", repo1.Name),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "full_name", repo1.FullName),
						resource.TestCheckResourceAttrSet("woodpecker_repository.test_repo", "avatar_url"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "default_branch", repo1.DefaultBranch),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "timeout", "30"),
						resource.TestCheckResourceAttr(
							"woodpecker_repository.test_repo",
							"visibility",
							woodpecker.VisibilityModePrivate.String(),
						),
						resource.TestCheckResourceAttr(
							"woodpecker_repository.test_repo",
							"is_private",
							strconv.FormatBool(repo1.Private),
						),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "trusted.network", "true"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "trusted.security", "false"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "trusted.volumes", "true"),
						resource.TestCheckResourceAttr(
							"woodpecker_repository.test_repo",
							"require_approval",
							woodpecker.ApprovalModeForks.String(),
						),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "is_active", "true"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "allow_pull_requests", "false"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "allow_deployments", "true"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "config_file", ".woodpecker2.yaml"),
						resource.TestCheckTypeSetElemAttr(
							"woodpecker_repository.test_repo",
							"cancel_previous_pipeline_events.*",
							woodpecker.EventTag,
						),
						resource.TestCheckTypeSetElemAttr(
							"woodpecker_repository.test_repo",
							"netrc_trusted_plugins.*",
							"woodpeckerci/plugin-docker-buildx",
						),
						resource.TestCheckTypeSetElemAttr(
							"woodpecker_repository.test_repo",
							"approval_allowed_users.*",
							repo1.Owner.UserName,
						),
					),
				},
				{ // import
					ResourceName:      "woodpecker_repository.test_repo",
					ImportState:       true,
					ImportStateId:     repo1.FullName,
					ImportStateVerify: true,
				},
				{ // replace repo
					Config: fmt.Sprintf(`
resource "woodpecker_repository" "test_repo" {
	full_name = "%s"
}
				`, repo2.FullName),
					ConfigPlanChecks: resource.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{
							plancheck.ExpectResourceAction("woodpecker_repository.test_repo", plancheck.ResourceActionReplace),
						},
					},
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("woodpecker_repository.test_repo", "id"),
						resource.TestCheckResourceAttrSet("woodpecker_repository.test_repo", "forge_id"),
						resource.TestCheckResourceAttr(
							"woodpecker_repository.test_repo",
							"forge_remote_id",
							strconv.FormatInt(repo2.ID, 10),
						),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "forge_url", repo2.HTMLURL),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "clone_url", repo2.CloneURL),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "owner", repo2.Owner.UserName),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "name", repo2.Name),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "full_name", repo2.FullName),
						resource.TestCheckResourceAttrSet("woodpecker_repository.test_repo", "avatar_url"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "default_branch", repo2.DefaultBranch),
						resource.TestCheckResourceAttrSet("woodpecker_repository.test_repo", "timeout"),
						resource.TestCheckResourceAttr(
							"woodpecker_repository.test_repo",
							"visibility",
							woodpecker.VisibilityModePublic.String(),
						),
						resource.TestCheckResourceAttr(
							"woodpecker_repository.test_repo",
							"is_private",
							strconv.FormatBool(repo2.Private),
						),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "trusted.network", "false"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "trusted.security", "false"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "trusted.volumes", "false"),
						resource.TestCheckResourceAttr(
							"woodpecker_repository.test_repo",
							"require_approval",
							woodpecker.ApprovalModeForks.String(),
						),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "is_active", "true"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "allow_pull_requests", "true"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "allow_deployments", "false"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "config_file", ""),
						resource.TestCheckTypeSetElemAttr(
							"woodpecker_repository.test_repo",
							"cancel_previous_pipeline_events.*",
							woodpecker.EventPush,
						),
						resource.TestCheckTypeSetElemAttr(
							"woodpecker_repository.test_repo",
							"cancel_previous_pipeline_events.*",
							woodpecker.EventPull,
						),
						resource.TestCheckNoResourceAttr("woodpecker_repository.test_repo", "netrc_trusted_plugins"),
					),
				},
			},
		})
	})

	t.Run("ERR: incorrect visibility value", func(t *testing.T) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`
resource "woodpecker_repository" "test_repo" {
	full_name = "%s"
	visibility = "asdf"
}
`, uuid.NewString()),
					ExpectError: regexp.MustCompile(`Attribute visibility value must be one of`),
				},
			},
		})
	})

	t.Run("ERR: incorrect require_approval value", func(t *testing.T) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`
resource "woodpecker_repository" "test_repo" {
	full_name = "%s"
	require_approval = "asdf"
}
`, uuid.NewString()),
					ExpectError: regexp.MustCompile(`Attribute require_approval value must be one of`),
				},
			},
		})
	})

	t.Run("ERR: incorrect cancel_previous_pipeline_events value", func(t *testing.T) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`
resource "woodpecker_repository" "test_repo" {
	full_name = "%s"
	cancel_previous_pipeline_events = ["asdf"]
}
`, uuid.NewString()),
					ExpectError: regexp.MustCompile(`Attribute cancel_previous_pipeline_events\[Value\("asdf"\)] value must be one`),
				},
			},
		})
	})

	t.Run("ERR: timeout needs to be >= 1", func(t *testing.T) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`
resource "woodpecker_repository" "test_repo" {
	full_name = "%s"
	timeout = 0
}
`, uuid.NewString()),
					ExpectError: regexp.MustCompile(`Attribute timeout value must be at least 1`),
				},
			},
		})
	})
}

func checkRepositoryResourceDestroy(names ...string) func(state *terraform.State) error {
	return func(_ *terraform.State) error {
		repos, err := woodpeckerClient.RepoList(woodpecker.RepoListOptions{All: true})
		if err != nil {
			return fmt.Errorf("couldn't list repos: %w", err)
		}

		if slices.ContainsFunc(repos, func(repo *woodpecker.Repo) bool {
			return slices.Contains(names, repo.FullName) && repo.IsActive
		}) {
			return errors.New("at least one of the repositories isn't inactive")
		}

		return nil
	}
}
