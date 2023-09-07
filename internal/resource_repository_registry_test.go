package internal_test

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/woodpecker-ci/woodpecker/woodpecker-go/woodpecker"
)

func TestRepositoryRegistryResource(t *testing.T) {
	t.Parallel()

	giteaRepo := createRepo(t)
	repo := activateRepo(t, giteaRepo)
	newRepo := activateRepo(t, createRepo(t))

	address := fmt.Sprintf("%s.localhost", uuid.NewString())
	newAddress := fmt.Sprintf("%s.localhost", uuid.NewString())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: checkRepositoryRegistryResourceDestroy(map[int64][]string{
			repo.ID:    {address, newAddress},
			newRepo.ID: {address, newAddress},
		}),
		Steps: []resource.TestStep{
			{ // create registry
				Config: fmt.Sprintf(`
resource "woodpecker_repository_registry" "test_registry" {
	repository_id = %d
	address = "%s"
	username = "test"
	password = "test"
}
`, repo.ID, address),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("woodpecker_repository_registry.test_registry", "id"),
					resource.TestCheckResourceAttr("woodpecker_repository_registry.test_registry", "repository_id", strconv.FormatInt(repo.ID, 10)),
					resource.TestCheckResourceAttr("woodpecker_repository_registry.test_registry", "address", address),
					resource.TestCheckResourceAttr("woodpecker_repository_registry.test_registry", "username", "test"),
					resource.TestCheckResourceAttr("woodpecker_repository_registry.test_registry", "password", "test"),
				),
			},
			{ // update registry
				Config: fmt.Sprintf(`
resource "woodpecker_repository_registry" "test_registry" {
	repository_id = %d
	address = "%s"
	username = "test2"
	password = "test2"
	email = "test@localhost"
}
`, repo.ID, address),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("woodpecker_repository_registry.test_registry", "id"),
					resource.TestCheckResourceAttr("woodpecker_repository_registry.test_registry", "repository_id", strconv.FormatInt(repo.ID, 10)),
					resource.TestCheckResourceAttr("woodpecker_repository_registry.test_registry", "address", address),
					resource.TestCheckResourceAttr("woodpecker_repository_registry.test_registry", "username", "test2"),
					resource.TestCheckResourceAttr("woodpecker_repository_registry.test_registry", "password", "test2"),
					resource.TestCheckResourceAttr("woodpecker_repository_registry.test_registry", "email", "test@localhost"),
				),
			},
			{ // update registry
				Config: fmt.Sprintf(`
resource "woodpecker_repository_registry" "test_registry" {
	repository_id = %d
	address = "%s"
	username = "test"
	password = "test"
}
//`, repo.ID, address),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("woodpecker_repository_registry.test_registry", "id"),
					resource.TestCheckResourceAttr("woodpecker_repository_registry.test_registry", "repository_id", strconv.FormatInt(repo.ID, 10)),
					resource.TestCheckResourceAttr("woodpecker_repository_registry.test_registry", "address", address),
					resource.TestCheckResourceAttr("woodpecker_repository_registry.test_registry", "username", "test"),
					resource.TestCheckResourceAttr("woodpecker_repository_registry.test_registry", "password", "test"),
					resource.TestCheckResourceAttr("woodpecker_repository_registry.test_registry", "email", "test@localhost"),
				),
			},
			{ // import
				ResourceName:            "woodpecker_repository_registry.test_registry",
				ImportState:             true,
				ImportStateId:           fmt.Sprintf("%d/%s", repo.ID, address),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
			{ // replace registry
				Config: fmt.Sprintf(`
resource "woodpecker_repository_registry" "test_registry" {
	repository_id = %d
	address = "%s"
	username = "test"
	password = "test"
}
`, repo.ID, newAddress),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("woodpecker_repository_registry.test_registry", plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("woodpecker_repository_registry.test_registry", "id"),
					resource.TestCheckResourceAttr("woodpecker_repository_registry.test_registry", "repository_id", strconv.FormatInt(repo.ID, 10)),
					resource.TestCheckResourceAttr("woodpecker_repository_registry.test_registry", "address", newAddress),
					resource.TestCheckResourceAttr("woodpecker_repository_registry.test_registry", "username", "test"),
					resource.TestCheckResourceAttr("woodpecker_repository_registry.test_registry", "password", "test"),
				),
			},
			{ // replace registry
				Config: fmt.Sprintf(`
resource "woodpecker_repository_registry" "test_registry" {
	repository_id = %d
	address = "%s"
	username = "test"
	password = "test"
}
`, newRepo.ID, newAddress),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("woodpecker_repository_registry.test_registry", plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("woodpecker_repository_registry.test_registry", "id"),
					resource.TestCheckResourceAttr("woodpecker_repository_registry.test_registry", "repository_id", strconv.FormatInt(newRepo.ID, 10)),
					resource.TestCheckResourceAttr("woodpecker_repository_registry.test_registry", "address", newAddress),
					resource.TestCheckResourceAttr("woodpecker_repository_registry.test_registry", "username", "test"),
					resource.TestCheckResourceAttr("woodpecker_repository_registry.test_registry", "password", "test"),
				),
			},
		},
	})
}

func checkRepositoryRegistryResourceDestroy(m map[int64][]string) func(state *terraform.State) error {
	return func(state *terraform.State) error {
		for repoID, addresses := range m {
			registries, err := woodpeckerClient.RegistryList(repoID)
			if err != nil {
				return fmt.Errorf("couldn't list registries: %w", err)
			}

			if slices.ContainsFunc(registries, func(registry *woodpecker.Registry) bool {
				return slices.Contains(addresses, registry.Address)
			}) {
				return errors.New("at least one of the created registries isn't deleted")
			}
		}

		return nil
	}
}
