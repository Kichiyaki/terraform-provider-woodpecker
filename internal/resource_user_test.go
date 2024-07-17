package internal_test

import (
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/Kichiyaki/terraform-provider-woodpecker/internal/woodpecker"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestUserResource(t *testing.T) {
	t.Parallel()

	login := uuid.NewString()
	newLogin := uuid.NewString()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             checkUserResourceDestroy(login, newLogin),
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
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "avatar_url", ""),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "is_active", "false"),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "is_admin", "false"),
				),
			},
			{ // update user
				Config: fmt.Sprintf(`
resource "woodpecker_user" "test_user" {
	login  = "%s"
	email  = "%s@localhost"
	avatar_url = "http://localhost/%s"
	is_admin  = true
}
`, login, login, login),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("woodpecker_user.test_user", "id"),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "login", login),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "email", login+"@localhost"),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "avatar_url", "http://localhost/"+login),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "is_active", "false"),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "is_admin", "true"),
				),
			},
			{ // update user
				Config: fmt.Sprintf(`
resource "woodpecker_user" "test_user" {
	login  = "%s"
	is_admin = false
}
`, login),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("woodpecker_user.test_user", "id"),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "login", login),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "email", login+"@localhost"),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "avatar_url", "http://localhost/"+login),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "is_active", "false"),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "is_admin", "false"),
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
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "avatar_url", ""),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "is_active", "false"),
					resource.TestCheckResourceAttr("woodpecker_user.test_user", "is_admin", "false"),
				),
			},
		},
	})
}

func checkUserResourceDestroy(logins ...string) func(state *terraform.State) error {
	return func(_ *terraform.State) error {
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
