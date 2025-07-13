package internal_test

import (
	"fmt"
	"testing"

	"github.com/Kichiyaki/terraform-provider-woodpecker/internal/woodpecker"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestOrgSecretDataSource(t *testing.T) {
	t.Parallel()

	org := createOrg(t)
	name := uuid.NewString()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "woodpecker_org_secret" "test_secret" {
	org_id = %d
	name = "%s"
	value = "test123"
	events = ["%s"]
}

data "woodpecker_org_secret" "test_secret" {
	org_id = woodpecker_org_secret.test_secret.org_id
	name = woodpecker_org_secret.test_secret.name
}
`, org.ID, name, woodpecker.EventPush),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.woodpecker_org_secret.test_secret", "id"),
					resource.TestCheckResourceAttrSet("data.woodpecker_org_secret.test_secret", "org_id"),
					resource.TestCheckResourceAttr("data.woodpecker_org_secret.test_secret", "name", name),
					resource.TestCheckTypeSetElemAttr(
						"data.woodpecker_org_secret.test_secret",
						"events.*",
						woodpecker.EventPush,
					),
				),
			},
		},
	})
}
