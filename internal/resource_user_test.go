package internal_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestResourceUser(t *testing.T) {
	t.Parallel()

	login := uuid.NewString()
	newLogin := uuid.NewString()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "woodpecker_user" "test_user" {
	login  = "%s"
}
`, login),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "login", login),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "email", ""),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "avatar", ""),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "active", "false"),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "admin", "false"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "woodpecker_user" "test_user" {
	login  = "%s"
	email  = "%s@localhost"
	avatar = "http://localhost/%s"
	admin  = true
}
`, login, login, login),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "login", login),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "email", login+"@localhost"),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "avatar", "http://localhost/"+login),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "active", "false"),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "admin", "true"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "woodpecker_user" "test_user" {
	login  = "%s"
}
`, login),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "login", login),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "email", login+"@localhost"),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "avatar", "http://localhost/"+login),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "active", "false"),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "admin", "true"),
				),
			},
			{
				ResourceName:      "woodpecker_user.test_user",
				ImportState:       true,
				ImportStateId:     login,
				ImportStateVerify: true,
			},
			{
				Config: fmt.Sprintf(`
resource "woodpecker_user" "test_user" {
	login  = "%s"
}
`, newLogin),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("woodpecker_user.test_user", plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "login", newLogin),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "email", ""),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "avatar", ""),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "active", "false"),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "admin", "false"),
				),
			},
		},
	})
}
