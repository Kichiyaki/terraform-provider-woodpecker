package internal_test

import (
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/woodpecker-ci/woodpecker/woodpecker-go/woodpecker"
)

func TestResourceUser(t *testing.T) {
	t.Parallel()

	login := uuid.NewString()
	newLogin := uuid.NewString()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckUserResourceDestroy(login, newLogin),
		Steps: []resource.TestStep{
			{ // create user
				Config: fmt.Sprintf(`
resource "woodpecker_user" "test_user" {
	login  = "%s"
}
`, login),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("woodpecker_user.test_user", "id"),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "login", login),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "email", ""),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "avatar", ""),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "active", "false"),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "admin", "false"),
				),
			},
			{ // update user
				Config: fmt.Sprintf(`
resource "woodpecker_user" "test_user" {
	login  = "%s"
	email  = "%s@localhost"
	avatar = "http://localhost/%s"
	admin  = true
}
`, login, login, login),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("woodpecker_user.test_user", "id"),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "login", login),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "email", login+"@localhost"),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "avatar", "http://localhost/"+login),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "active", "false"),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "admin", "true"),
				),
			},
			{ // fields shouldn't be overridden
				Config: fmt.Sprintf(`
resource "woodpecker_user" "test_user" {
	login  = "%s"
}
`, login),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("woodpecker_user.test_user", "id"),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "login", login),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "email", login+"@localhost"),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "avatar", "http://localhost/"+login),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "active", "false"),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "admin", "true"),
				),
			},
			{ // import
				ResourceName:      "woodpecker_user.test_user",
				ImportState:       true,
				ImportStateId:     login,
				ImportStateVerify: true,
			},
			{ // replace user
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
					resource.TestCheckResourceAttrSet("woodpecker_user.test_user", "id"),
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

func testAccCheckUserResourceDestroy(logins ...string) func(state *terraform.State) error {
	return func(state *terraform.State) error {
		users, err := woodpeckerClient.UserList()
		if err != nil {
			return fmt.Errorf("couldn't list users: %w", err)
		}

		if slices.ContainsFunc(users, func(user *woodpecker.User) bool {
			return slices.Contains(logins, user.Login)
		}) {
			return errors.New("at least one of the created users isn't deleted")
		}

		return nil
	}
}
