package internal_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestRepositoryRegistryDataSource(t *testing.T) {
	t.Parallel()

	repo := createRepo(t)
	address := uuid.NewString()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "woodpecker_repository" "test_repo" {
	full_name = "%s"
}

resource "woodpecker_repository_registry" "test_registry" {
	repository_id = woodpecker_repository.test_repo.id
	address = "%s"
	username = "test"
	password = "test"
}

data "woodpecker_repository_registry" "test_registry" {
	repository_id = woodpecker_repository_registry.test_registry.repository_id
	address = woodpecker_repository_registry.test_registry.address
}
`, repo.FullName, address),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.woodpecker_repository_registry.test_registry", "id"),
					resource.TestCheckResourceAttrSet("data.woodpecker_repository_registry.test_registry", "repository_id"),
					resource.TestCheckResourceAttr("data.woodpecker_repository_registry.test_registry", "address", address),
					resource.TestCheckResourceAttr("data.woodpecker_repository_registry.test_registry", "username", "test"),
				),
			},
		},
	})
}
