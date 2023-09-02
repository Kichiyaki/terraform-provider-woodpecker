package internal_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestSecretDataSource(t *testing.T) {
	t.Parallel()

	name := uuid.NewString()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "woodpecker_secret" "test_secret" {
	name = "%s"
	value = "test123"
	events = ["push"]
}

data "woodpecker_secret" "test_secret" {
	name = woodpecker_secret.test_secret.name
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.woodpecker_secret.test_secret", "id"),
					resource.TestCheckResourceAttr("data.woodpecker_secret.test_secret", "name", name),
					resource.TestCheckTypeSetElemAttr("data.woodpecker_secret.test_secret", "events.*", "push"),
					resource.TestCheckResourceAttr("data.woodpecker_secret.test_secret", "plugins_only", "false"),
				),
			},
		},
	})
}
