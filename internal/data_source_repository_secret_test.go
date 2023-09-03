package internal_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestRepositorySecretDataSource(t *testing.T) {
	t.Parallel()

	repo := createRepo(t)
	name := uuid.NewString()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "woodpecker_repository" "test_repo" {
	full_name = "%s"
}

resource "woodpecker_repository_secret" "test_secret" {
	repository_id = woodpecker_repository.test_repo.id
	name = "%s"
	value = "test123"
	events = ["push"]
}

data "woodpecker_repository_secret" "test_secret" {
	repository_id = woodpecker_repository_secret.test_secret.repository_id
	name = woodpecker_repository_secret.test_secret.name
}
`, repo.FullName, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.woodpecker_repository_secret.test_secret", "id"),
					resource.TestCheckResourceAttrSet("data.woodpecker_repository_secret.test_secret", "repository_id"),
					resource.TestCheckResourceAttr("data.woodpecker_repository_secret.test_secret", "name", name),
					resource.TestCheckTypeSetElemAttr("data.woodpecker_repository_secret.test_secret", "events.*", "push"),
					resource.TestCheckResourceAttr("data.woodpecker_repository_secret.test_secret", "plugins_only", "false"),
				),
			},
		},
	})
}
