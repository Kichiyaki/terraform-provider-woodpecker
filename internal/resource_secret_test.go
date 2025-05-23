package internal_test

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"testing"

	"github.com/Kichiyaki/terraform-provider-woodpecker/internal/woodpecker"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestSecretResource(t *testing.T) {
	t.Parallel()

	t.Run("OK", func(t *testing.T) {
		t.Parallel()

		name := uuid.NewString()
		newName := uuid.NewString()

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			CheckDestroy:             checkSecretResourceDestroy(name, newName),
			Steps: []resource.TestStep{
				{ // create secret
					Config: fmt.Sprintf(`
resource "woodpecker_secret" "test_secret" {
	name = "%s"
	value = "test123"
	events = ["%s"]
}
`, name, woodpecker.EventPush),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("woodpecker_secret.test_secret", "id"),
						resource.TestCheckResourceAttr("woodpecker_secret.test_secret", "name", name),
						resource.TestCheckResourceAttr("woodpecker_secret.test_secret", "value", "test123"),
						resource.TestCheckTypeSetElemAttr("woodpecker_secret.test_secret", "events.*", woodpecker.EventPush),
					),
				},
				{ // update secret
					Config: fmt.Sprintf(`
resource "woodpecker_secret" "test_secret" {
	name = "%s"
	value = "test123123"
	events = ["%s", "%s"]
}
`, name, woodpecker.EventPush, woodpecker.EventDeploy),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("woodpecker_secret.test_secret", "id"),
						resource.TestCheckResourceAttr("woodpecker_secret.test_secret", "name", name),
						resource.TestCheckResourceAttr("woodpecker_secret.test_secret", "value", "test123123"),
						resource.TestCheckTypeSetElemAttr("woodpecker_secret.test_secret", "events.*", woodpecker.EventPush),
						resource.TestCheckTypeSetElemAttr("woodpecker_secret.test_secret", "events.*", woodpecker.EventDeploy),
					),
				},
				{ // update secret
					Config: fmt.Sprintf(`
resource "woodpecker_secret" "test_secret" {
	name = "%s"
	value = "test123123"
	events = ["%s", "%s", "%s"]
	images = ["testimage"]
}
`, name, woodpecker.EventPush, woodpecker.EventDeploy, woodpecker.EventCron),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("woodpecker_secret.test_secret", "id"),
						resource.TestCheckResourceAttr("woodpecker_secret.test_secret", "name", name),
						resource.TestCheckResourceAttr("woodpecker_secret.test_secret", "value", "test123123"),
						resource.TestCheckTypeSetElemAttr("woodpecker_secret.test_secret", "events.*", woodpecker.EventPush),
						resource.TestCheckTypeSetElemAttr("woodpecker_secret.test_secret", "events.*", woodpecker.EventDeploy),
						resource.TestCheckTypeSetElemAttr("woodpecker_secret.test_secret", "events.*", woodpecker.EventCron),
						resource.TestCheckTypeSetElemAttr("woodpecker_secret.test_secret", "images.*", "testimage"),
					),
				},
				{ // nothing to update
					Config: fmt.Sprintf(`
resource "woodpecker_secret" "test_secret" {
	name = "%s"
	value = "test123123"
	events = ["%s", "%s", "%s"]
}
`, name, woodpecker.EventPush, woodpecker.EventDeploy, woodpecker.EventCron),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("woodpecker_secret.test_secret", "id"),
						resource.TestCheckResourceAttr("woodpecker_secret.test_secret", "name", name),
						resource.TestCheckResourceAttr("woodpecker_secret.test_secret", "value", "test123123"),
						resource.TestCheckTypeSetElemAttr("woodpecker_secret.test_secret", "events.*", woodpecker.EventPush),
						resource.TestCheckTypeSetElemAttr("woodpecker_secret.test_secret", "events.*", woodpecker.EventDeploy),
						resource.TestCheckTypeSetElemAttr("woodpecker_secret.test_secret", "events.*", woodpecker.EventCron),
						resource.TestCheckTypeSetElemAttr("woodpecker_secret.test_secret", "images.*", "testimage"),
					),
				},
				{ // import
					ResourceName:            "woodpecker_secret.test_secret",
					ImportState:             true,
					ImportStateId:           name,
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"value"},
				},
				{ // replace secret
					Config: fmt.Sprintf(`
resource "woodpecker_secret" "test_secret" {
	name = "%s"
	value = "test123New"
	events = ["%s"]
}
`, newName, woodpecker.EventPush),
					ConfigPlanChecks: resource.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{
							plancheck.ExpectResourceAction("woodpecker_secret.test_secret", plancheck.ResourceActionReplace),
						},
					},
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("woodpecker_secret.test_secret", "id"),
						resource.TestCheckResourceAttr("woodpecker_secret.test_secret", "name", newName),
						resource.TestCheckResourceAttr("woodpecker_secret.test_secret", "value", "test123New"),
						resource.TestCheckTypeSetElemAttr("woodpecker_secret.test_secret", "events.*", woodpecker.EventPush),
					),
				},
			},
		})
	})

	t.Run("ERR: incorrect event value", func(t *testing.T) {
		t.Parallel()

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`
resource "woodpecker_secret" "test_secret" {
	name = "%s"
	value = "test123"
	events = ["random"]
}
`, uuid.NewString()),
					ExpectError: regexp.MustCompile(`Attribute events\[Value\("random"\)] value must be one of`),
				},
			},
		})
	})
}

func checkSecretResourceDestroy(names ...string) func(state *terraform.State) error {
	return func(_ *terraform.State) error {
		secrets, err := woodpeckerClient.GlobalSecretList(woodpecker.SecretListOptions{})
		if err != nil {
			return fmt.Errorf("couldn't list secrets: %w", err)
		}

		if slices.ContainsFunc(secrets, func(secret *woodpecker.Secret) bool {
			return slices.Contains(names, secret.Name)
		}) {
			return errors.New("at least one of the created secrets isn't deleted")
		}

		return nil
	}
}
