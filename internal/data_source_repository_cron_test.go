package internal_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestRepositoryCronDataSource(t *testing.T) {
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

resource "woodpecker_repository_cron" "test_cron" {
	repository_id = woodpecker_repository.test_repo.id
	name = "%s"
	schedule = "@daily"
}

data "woodpecker_repository_cron" "test_cron" {
	repository_id = woodpecker_repository_cron.test_cron.repository_id
	id = woodpecker_repository_cron.test_cron.id
}
`, repo.FullName, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.woodpecker_repository_cron.test_cron", "id"),
					resource.TestCheckResourceAttrSet("data.woodpecker_repository_cron.test_cron", "repository_id"),
					resource.TestCheckResourceAttr("data.woodpecker_repository_cron.test_cron", "name", name),
					resource.TestCheckResourceAttr("data.woodpecker_repository_cron.test_cron", "schedule", "@daily"),
					resource.TestCheckResourceAttr("data.woodpecker_repository_cron.test_cron", "branch", ""),
					resource.TestCheckResourceAttrSet("data.woodpecker_repository_cron.test_cron", "created_at"),
					resource.TestCheckResourceAttrSet("data.woodpecker_repository_cron.test_cron", "creator_id"),
				),
			},
		},
	})
}
