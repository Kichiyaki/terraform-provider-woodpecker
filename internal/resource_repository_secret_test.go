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

func TestRepositorySecretResource(t *testing.T) {
	t.Parallel()

	t.Run("OK", func(t *testing.T) {
		t.Parallel()

		repo := activateRepo(t, createRepo(t))
		newRepo := activateRepo(t, createRepo(t))

		name := uuid.NewString()
		newName := uuid.NewString()

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			CheckDestroy: checkRepositorySecretResourceDestroy(map[int64][]string{
				repo.ID:    {name, newName},
				newRepo.ID: {name, newName},
			}),
			Steps: []resource.TestStep{
				{ // create secret
					Config: fmt.Sprintf(`
resource "woodpecker_repository_secret" "test_secret" {
	repository_id = %d
	name = "%s"
	value = "test123"
	events = ["push"]
}
`, repo.ID, name),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("woodpecker_repository_secret.test_secret", "id"),
						resource.TestCheckResourceAttr(
							"woodpecker_repository_secret.test_secret",
							"repository_id",
							strconv.FormatInt(repo.ID, 10),
						),
						resource.TestCheckResourceAttr("woodpecker_repository_secret.test_secret", "name", name),
						resource.TestCheckResourceAttr("woodpecker_repository_secret.test_secret", "value", "test123"),
						resource.TestCheckTypeSetElemAttr("woodpecker_repository_secret.test_secret", "events.*", "push"),
					),
				},
				{ // update secret
					Config: fmt.Sprintf(`
resource "woodpecker_repository_secret" "test_secret" {
	repository_id = %d
	name = "%s"
	value = "test123123"
	events = ["push", "deployment"]
	images = ["testimage"]
}
`, repo.ID, name),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("woodpecker_repository_secret.test_secret", "id"),
						resource.TestCheckResourceAttr(
							"woodpecker_repository_secret.test_secret",
							"repository_id",
							strconv.FormatInt(repo.ID, 10),
						),
						resource.TestCheckResourceAttr("woodpecker_repository_secret.test_secret", "name", name),
						resource.TestCheckResourceAttr("woodpecker_repository_secret.test_secret", "value", "test123123"),
						resource.TestCheckTypeSetElemAttr("woodpecker_repository_secret.test_secret", "events.*", "push"),
						resource.TestCheckTypeSetElemAttr("woodpecker_repository_secret.test_secret", "events.*", "deployment"),
						resource.TestCheckTypeSetElemAttr("woodpecker_repository_secret.test_secret", "images.*", "testimage"),
					),
				},
				{ // update secret
					Config: fmt.Sprintf(`
resource "woodpecker_repository_secret" "test_secret" {
	repository_id = %d
	name = "%s"
	value = "test123123"
	events = ["push", "deployment", "cron"]
}
//`, repo.ID, name),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("woodpecker_repository_secret.test_secret", "id"),
						resource.TestCheckResourceAttr(
							"woodpecker_repository_secret.test_secret",
							"repository_id",
							strconv.FormatInt(repo.ID, 10),
						),
						resource.TestCheckResourceAttr("woodpecker_repository_secret.test_secret", "name", name),
						resource.TestCheckResourceAttr("woodpecker_repository_secret.test_secret", "value", "test123123"),
						resource.TestCheckTypeSetElemAttr("woodpecker_repository_secret.test_secret", "events.*", "push"),
						resource.TestCheckTypeSetElemAttr("woodpecker_repository_secret.test_secret", "events.*", "deployment"),
						resource.TestCheckTypeSetElemAttr("woodpecker_repository_secret.test_secret", "events.*", "cron"),
						resource.TestCheckTypeSetElemAttr("woodpecker_repository_secret.test_secret", "images.*", "testimage"),
					),
				},
				{ // import
					ResourceName:            "woodpecker_repository_secret.test_secret",
					ImportState:             true,
					ImportStateId:           fmt.Sprintf("%d/%s", repo.ID, name),
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"value"},
				},
				{ // replace secret (new name)
					Config: fmt.Sprintf(`
resource "woodpecker_repository_secret" "test_secret" {
	repository_id = %d
	name = "%s"
	value = "test123New"
	events = ["push"]
}
`, repo.ID, newName),
					ConfigPlanChecks: resource.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{
							plancheck.ExpectResourceAction("woodpecker_repository_secret.test_secret", plancheck.ResourceActionReplace),
						},
					},
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("woodpecker_repository_secret.test_secret", "id"),
						resource.TestCheckResourceAttr(
							"woodpecker_repository_secret.test_secret",
							"repository_id",
							strconv.FormatInt(repo.ID, 10),
						),
						resource.TestCheckResourceAttr("woodpecker_repository_secret.test_secret", "name", newName),
						resource.TestCheckResourceAttr("woodpecker_repository_secret.test_secret", "value", "test123New"),
						resource.TestCheckTypeSetElemAttr("woodpecker_repository_secret.test_secret", "events.*", "push"),
					),
				},
				{ // replace secret (new repo id)
					Config: fmt.Sprintf(`
resource "woodpecker_repository_secret" "test_secret" {
	repository_id = %d
	name = "%s"
	value = "test123New"
	events = ["push"]
}
`, newRepo.ID, newName),
					ConfigPlanChecks: resource.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{
							plancheck.ExpectResourceAction("woodpecker_repository_secret.test_secret", plancheck.ResourceActionReplace),
						},
					},
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("woodpecker_repository_secret.test_secret", "id"),
						resource.TestCheckResourceAttr(
							"woodpecker_repository_secret.test_secret",
							"repository_id",
							strconv.FormatInt(newRepo.ID, 10),
						),
						resource.TestCheckResourceAttr("woodpecker_repository_secret.test_secret", "name", newName),
						resource.TestCheckResourceAttr("woodpecker_repository_secret.test_secret", "value", "test123New"),
						resource.TestCheckTypeSetElemAttr("woodpecker_repository_secret.test_secret", "events.*", "push"),
					),
				},
			},
		})
	})

	t.Run("ERR: incorrect event value", func(t *testing.T) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`
resource "woodpecker_repository_secret" "test_secret" {
	repository_id = 123
	name = "%s"
	value = "test123"
	events = ["random"]
}
`, uuid.NewString()),
					ExpectError: regexp.MustCompile(`Attribute events\[Value\("random"\)] value must be one of`),
				},
			},
		})
	})
}

func checkRepositorySecretResourceDestroy(m map[int64][]string) func(state *terraform.State) error {
	return func(state *terraform.State) error {
		for repoID, names := range m {
			secrets, err := woodpeckerClient.SecretList(repoID)
			if err != nil {
				return fmt.Errorf("couldn't list secrets: %w", err)
			}

			if slices.ContainsFunc(secrets, func(secret *woodpecker.Secret) bool {
				return slices.Contains(names, secret.Name)
			}) {
				return errors.New("at least one of the created secrets isn't deleted")
			}
		}

		return nil
	}
}
