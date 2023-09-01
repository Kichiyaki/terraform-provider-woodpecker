package internal_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestDataSourceUser(t *testing.T) {
	t.Parallel()

	user, _ := woodpeckerClient.Self()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: buildDataSourceUserConfig("current", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.woodpecker_user.current", "id", strconv.FormatInt(user.ID, 10)),
					resource.TestCheckResourceAttr("data.woodpecker_user.current", "login", user.Login),
					resource.TestCheckResourceAttr("data.woodpecker_user.current", "email", user.Email),
					resource.TestCheckResourceAttr("data.woodpecker_user.current", "avatar_url", user.Avatar),
					resource.TestCheckResourceAttr("data.woodpecker_user.current", "is_admin", strconv.FormatBool(user.Admin)),
					resource.TestCheckResourceAttr("data.woodpecker_user.current", "is_active", strconv.FormatBool(user.Active)),
				),
			},
			{
				Config: buildDataSourceUserConfig("user", user.Login),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.woodpecker_user.user", "id", strconv.FormatInt(user.ID, 10)),
					resource.TestCheckResourceAttr("data.woodpecker_user.user", "login", user.Login),
					resource.TestCheckResourceAttr("data.woodpecker_user.user", "email", user.Email),
					resource.TestCheckResourceAttr("data.woodpecker_user.user", "avatar_url", user.Avatar),
					resource.TestCheckResourceAttr("data.woodpecker_user.user", "is_admin", strconv.FormatBool(user.Admin)),
					resource.TestCheckResourceAttr("data.woodpecker_user.user", "is_active", strconv.FormatBool(user.Active)),
				),
			},
		},
	})
}

func buildDataSourceUserConfig(name, login string) string {
	return fmt.Sprintf(`
data "woodpecker_user" "%s" {
	login = "%s"
}
`, name, login)
}
