package internal_test

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"testing"

	"github.com/Kichiyaki/terraform-provider-woodpecker/internal/woodpecker"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestRepositoryCronResource(t *testing.T) {
	t.Parallel()

	giteaRepo := createRepo(t)
	repo := activateRepo(t, giteaRepo)
	branch := createBranch(t, giteaRepo)
	newRepo := activateRepo(t, createRepo(t))

	name := uuid.NewString()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: checkRepositoryCronResourceDestroy(map[int64][]string{
			repo.ID:    {name},
			newRepo.ID: {name},
		}),
		Steps: []resource.TestStep{
			{ // create cron
				Config: fmt.Sprintf(`
resource "woodpecker_repository_cron" "test_cron" {
	repository_id = %d
	name = "%s"
	schedule = "@daily"
}
`, repo.ID, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("woodpecker_repository_cron.test_cron", "id"),
					resource.TestCheckResourceAttr(
						"woodpecker_repository_cron.test_cron",
						"repository_id",
						strconv.FormatInt(repo.ID, 10),
					),
					resource.TestCheckResourceAttr("woodpecker_repository_cron.test_cron", "name", name),
					resource.TestCheckResourceAttr("woodpecker_repository_cron.test_cron", "schedule", "@daily"),
					resource.TestCheckResourceAttr("woodpecker_repository_cron.test_cron", "branch", ""),
					resource.TestCheckResourceAttrSet("woodpecker_repository_cron.test_cron", "created_at"),
					resource.TestCheckResourceAttrSet("woodpecker_repository_cron.test_cron", "creator_id"),
				),
			},
			{ // update cron
				Config: fmt.Sprintf(`
resource "woodpecker_repository_cron" "test_cron" {
	repository_id = %d
	name = "%s"
	schedule = "@every 5m"
	branch = "%s"
}
`, repo.ID, name, branch.Name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("woodpecker_repository_cron.test_cron", "id"),
					resource.TestCheckResourceAttr(
						"woodpecker_repository_cron.test_cron",
						"repository_id",
						strconv.FormatInt(repo.ID, 10),
					),
					resource.TestCheckResourceAttr("woodpecker_repository_cron.test_cron", "name", name),
					resource.TestCheckResourceAttr("woodpecker_repository_cron.test_cron", "schedule", "@every 5m"),
					resource.TestCheckResourceAttr("woodpecker_repository_cron.test_cron", "branch", branch.Name),
					resource.TestCheckResourceAttrSet("woodpecker_repository_cron.test_cron", "created_at"),
					resource.TestCheckResourceAttrSet("woodpecker_repository_cron.test_cron", "creator_id"),
				),
			},
			{ // update cron
				Config: fmt.Sprintf(`
resource "woodpecker_repository_cron" "test_cron" {
	repository_id = %d
	name = "%s"
	schedule = "@daily"
}
//`, repo.ID, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("woodpecker_repository_cron.test_cron", "id"),
					resource.TestCheckResourceAttr(
						"woodpecker_repository_cron.test_cron",
						"repository_id",
						strconv.FormatInt(repo.ID, 10),
					),
					resource.TestCheckResourceAttr("woodpecker_repository_cron.test_cron", "name", name),
					resource.TestCheckResourceAttr("woodpecker_repository_cron.test_cron", "schedule", "@daily"),
					resource.TestCheckResourceAttr("woodpecker_repository_cron.test_cron", "branch", branch.Name),
					resource.TestCheckResourceAttrSet("woodpecker_repository_cron.test_cron", "created_at"),
					resource.TestCheckResourceAttrSet("woodpecker_repository_cron.test_cron", "creator_id"),
				),
			},
			{ // import
				ResourceName:        "woodpecker_repository_cron.test_cron",
				ImportState:         true,
				ImportStateIdPrefix: strconv.FormatInt(repo.ID, 10) + "/",
				ImportStateVerify:   true,
			},
			{ // replace cron
				Config: fmt.Sprintf(`
resource "woodpecker_repository_cron" "test_cron" {
	repository_id = %d
	name = "%s"
	schedule = "@daily"
}
`, newRepo.ID, name),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("woodpecker_repository_cron.test_cron", plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("woodpecker_repository_cron.test_cron", "id"),
					resource.TestCheckResourceAttr(
						"woodpecker_repository_cron.test_cron",
						"repository_id",
						strconv.FormatInt(newRepo.ID, 10),
					),
					resource.TestCheckResourceAttr("woodpecker_repository_cron.test_cron", "name", name),
					resource.TestCheckResourceAttr("woodpecker_repository_cron.test_cron", "schedule", "@daily"),
					resource.TestCheckResourceAttr("woodpecker_repository_cron.test_cron", "branch", ""),
					resource.TestCheckResourceAttrSet("woodpecker_repository_cron.test_cron", "created_at"),
					resource.TestCheckResourceAttrSet("woodpecker_repository_cron.test_cron", "creator_id"),
				),
			},
		},
	})
}

func checkRepositoryCronResourceDestroy(m map[int64][]string) func(state *terraform.State) error {
	return func(_ *terraform.State) error {
		for repoID, names := range m {
			crons, err := woodpeckerClient.CronList(repoID)
			if err != nil {
				return fmt.Errorf("couldn't list cron jobs: %w", err)
			}

			if slices.ContainsFunc(crons, func(cron *woodpecker.Cron) bool {
				return slices.Contains(names, cron.Name)
			}) {
				return errors.New("at least one of the created cron jobs isn't deleted")
			}
		}

		return nil
	}
}
