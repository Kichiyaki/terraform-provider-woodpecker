package internal_test

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"testing"

	"code.gitea.io/sdk/gitea"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/woodpecker-ci/woodpecker/woodpecker-go/woodpecker"
)

func TestResourceRepository(t *testing.T) {
	t.Parallel()

	t.Run("OK", func(t *testing.T) {
		t.Parallel()

		repo1, repo2 := createRepo(t), createRepo(t)

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			CheckDestroy:             testAccCheckRepositoryResourceDestroy(repo1.FullName, repo2.FullName),
			Steps: []resource.TestStep{
				{ // create repo
					Config: fmt.Sprintf(`
resource "woodpecker_repository" "test_repo" {
	full_name = "%s"
	is_trusted = true
	visibility = "public"
}
`, repo1.FullName),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("woodpecker_repository.test_repo", "id"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "forge_remote_id", strconv.FormatInt(repo1.ID, 10)),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "name", repo1.Name),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "owner", repo1.Owner.UserName),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "full_name", repo1.FullName),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "url", repo1.HTMLURL),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "clone_url", repo1.CloneURL),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "default_branch", repo1.DefaultBranch),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "scm", "git"),
						resource.TestCheckResourceAttrSet("woodpecker_repository.test_repo", "timeout"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "visibility", "public"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "is_private", strconv.FormatBool(repo1.Private)),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "is_trusted", "true"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "is_gated", "false"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "allow_pull_requests", "true"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "config_file", ""),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "netrc_only_trusted", "true"),
					),
				},
				{ // update repo
					Config: fmt.Sprintf(`
resource "woodpecker_repository" "test_repo" {
	full_name = "%s"
	is_trusted = false
	visibility = "private"
	is_gated = true
	timeout = 30
	config_file = ".woodpecker2.yaml"
}
`, repo1.FullName),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("woodpecker_repository.test_repo", "id"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "forge_remote_id", strconv.FormatInt(repo1.ID, 10)),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "name", repo1.Name),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "owner", repo1.Owner.UserName),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "full_name", repo1.FullName),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "url", repo1.HTMLURL),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "clone_url", repo1.CloneURL),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "default_branch", repo1.DefaultBranch),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "scm", "git"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "timeout", "30"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "visibility", "private"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "is_private", strconv.FormatBool(repo1.Private)),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "is_trusted", "false"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "is_gated", "true"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "allow_pull_requests", "true"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "config_file", ".woodpecker2.yaml"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "netrc_only_trusted", "true"),
					),
				},
				{ // update repo
					Config: fmt.Sprintf(`
resource "woodpecker_repository" "test_repo" {
	full_name = "%s"
	timeout = 15
}
				//`, repo1.FullName),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("woodpecker_repository.test_repo", "id"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "forge_remote_id", strconv.FormatInt(repo1.ID, 10)),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "name", repo1.Name),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "owner", repo1.Owner.UserName),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "full_name", repo1.FullName),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "url", repo1.HTMLURL),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "clone_url", repo1.CloneURL),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "default_branch", repo1.DefaultBranch),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "scm", "git"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "timeout", "15"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "visibility", "private"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "is_private", strconv.FormatBool(repo1.Private)),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "is_trusted", "false"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "is_gated", "true"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "allow_pull_requests", "true"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "config_file", ".woodpecker2.yaml"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "netrc_only_trusted", "true"),
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
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "forge_remote_id", strconv.FormatInt(repo2.ID, 10)),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "name", repo2.Name),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "owner", repo2.Owner.UserName),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "full_name", repo2.FullName),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "url", repo2.HTMLURL),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "clone_url", repo2.CloneURL),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "default_branch", repo2.DefaultBranch),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "scm", "git"),
						resource.TestCheckResourceAttrSet("woodpecker_repository.test_repo", "timeout"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "visibility", "public"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "is_private", strconv.FormatBool(repo2.Private)),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "is_trusted", "false"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "is_gated", "false"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "allow_pull_requests", "true"),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "config_file", ""),
						resource.TestCheckResourceAttr("woodpecker_repository.test_repo", "netrc_only_trusted", "true"),
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

func testAccCheckRepositoryResourceDestroy(names ...string) func(state *terraform.State) error {
	return func(state *terraform.State) error {
		repos, err := woodpeckerClient.RepoListOpts(true, true)
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

func createRepo(tb testing.TB) *gitea.Repository {
	tb.Helper()

	repo, _, err := giteaClient.CreateRepo(gitea.CreateRepoOption{
		Name:          uuid.NewString(),
		Description:   uuid.NewString(),
		Private:       false,
		AutoInit:      true,
		Template:      false,
		License:       "MIT",
		Readme:        "Default",
		DefaultBranch: "master",
	})
	if err != nil {
		tb.Fatalf("got unexpected error while creating repo: %s", err)
	}

	return repo
}
