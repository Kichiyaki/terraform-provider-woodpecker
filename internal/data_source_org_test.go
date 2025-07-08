package internal_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestOrgDataSource(t *testing.T) {
	t.Parallel()

	repo1 := createOrgRepo(t)
	activateRepo(t, repo1)
	repo2 := createRepo(t)
	activateRepo(t, repo2)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "woodpecker_org" "test_org" {
	name = "%s"
}
`, repo1.Owner.UserName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.woodpecker_org.test_org", "id"),
					resource.TestCheckResourceAttrSet("data.woodpecker_org.test_org", "forge_id"),
					resource.TestCheckResourceAttr("data.woodpecker_org.test_org", "name", repo1.Owner.UserName),
					resource.TestCheckResourceAttr("data.woodpecker_org.test_org", "is_user", "false"),
				),
			},
			{
				Config: fmt.Sprintf(`
data "woodpecker_org" "test_org" {
	name = "%s"
}
`, repo2.Owner.UserName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.woodpecker_org.test_org", "id"),
					resource.TestCheckResourceAttrSet("data.woodpecker_org.test_org", "forge_id"),
					resource.TestCheckResourceAttr("data.woodpecker_org.test_org", "name", repo2.Owner.UserName),
					resource.TestCheckResourceAttr("data.woodpecker_org.test_org", "is_user", "true"),
				),
			},
		},
	})
}
