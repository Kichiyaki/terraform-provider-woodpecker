package internal_test

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"testing"

	"github.com/Kichiyaki/terraform-provider-woodpecker/internal/woodpecker"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestOrgSecretResource(t *testing.T) {
	t.Parallel()

	t.Run("OK", func(t *testing.T) {
		t.Parallel()

		org := createOrg(t)
		newOrg := createOrg(t)

		name := uuid.NewString()
		newName := uuid.NewString()

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			CheckDestroy: checkOrgSecretResourceDestroy(map[int64][]string{
				org.ID:    {name, newName},
				newOrg.ID: {name, newName},
			}),
			Steps: []resource.TestStep{
				{ // create secret
					Config: fmt.Sprintf(`
resource "woodpecker_org_secret" "test_secret" {
	org_id = %d
	name = "%s"
	value = "test123"
	events = ["%s"]
}
`, org.ID, name, woodpecker.EventPush),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("woodpecker_org_secret.test_secret", "id"),
						resource.TestCheckResourceAttr(
							"woodpecker_org_secret.test_secret",
							"org_id",
							strconv.FormatInt(org.ID, 10),
						),
						resource.TestCheckResourceAttr("woodpecker_org_secret.test_secret", "name", name),
						resource.TestCheckResourceAttr("woodpecker_org_secret.test_secret", "value", "test123"),
						resource.TestCheckTypeSetElemAttr("woodpecker_org_secret.test_secret", "events.*", woodpecker.EventPush),
					),
				},
				{ // update secret
					Config: fmt.Sprintf(`
resource "woodpecker_org_secret" "test_secret" {
	org_id = %d
	name = "%s"
	value = "test123123"
	events = ["%s", "%s", "%s", "%s"]
	images = ["testimage"]
}
`, org.ID, name, woodpecker.EventPush, woodpecker.EventDeploy, woodpecker.EventRelease, woodpecker.EventPullClosed),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("woodpecker_org_secret.test_secret", "id"),
						resource.TestCheckResourceAttr(
							"woodpecker_org_secret.test_secret",
							"org_id",
							strconv.FormatInt(org.ID, 10),
						),
						resource.TestCheckResourceAttr("woodpecker_org_secret.test_secret", "name", name),
						resource.TestCheckResourceAttr("woodpecker_org_secret.test_secret", "value", "test123123"),
						resource.TestCheckTypeSetElemAttr("woodpecker_org_secret.test_secret", "events.*", woodpecker.EventPush),
						resource.TestCheckTypeSetElemAttr("woodpecker_org_secret.test_secret", "events.*", woodpecker.EventDeploy),
						resource.TestCheckTypeSetElemAttr(
							"woodpecker_org_secret.test_secret",
							"events.*",
							woodpecker.EventRelease,
						),
						resource.TestCheckTypeSetElemAttr(
							"woodpecker_org_secret.test_secret",
							"events.*",
							woodpecker.EventPullClosed,
						),
						resource.TestCheckTypeSetElemAttr("woodpecker_org_secret.test_secret", "images.*", "testimage"),
					),
				},
				{ // update secret
					Config: fmt.Sprintf(`
resource "woodpecker_org_secret" "test_secret" {
	org_id = %d
	name = "%s"
	value = "test123123"
	events = ["%s", "%s", "%s"]
}
//`, org.ID, name, woodpecker.EventPush, woodpecker.EventDeploy, woodpecker.EventCron),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("woodpecker_org_secret.test_secret", "id"),
						resource.TestCheckResourceAttr(
							"woodpecker_org_secret.test_secret",
							"org_id",
							strconv.FormatInt(org.ID, 10),
						),
						resource.TestCheckResourceAttr("woodpecker_org_secret.test_secret", "name", name),
						resource.TestCheckResourceAttr("woodpecker_org_secret.test_secret", "value", "test123123"),
						resource.TestCheckTypeSetElemAttr("woodpecker_org_secret.test_secret", "events.*", woodpecker.EventPush),
						resource.TestCheckTypeSetElemAttr("woodpecker_org_secret.test_secret", "events.*", woodpecker.EventDeploy),
						resource.TestCheckTypeSetElemAttr("woodpecker_org_secret.test_secret", "events.*", woodpecker.EventCron),
						resource.TestCheckTypeSetElemAttr("woodpecker_org_secret.test_secret", "images.*", "testimage"),
					),
				},
				{ // import
					ResourceName:            "woodpecker_org_secret.test_secret",
					ImportState:             true,
					ImportStateId:           fmt.Sprintf("%d/%s", org.ID, name),
					ImportStateVerify:       true,
					ImportStateVerifyIgnore: []string{"value"},
				},
				{ // replace secret (new name)
					Config: fmt.Sprintf(`
resource "woodpecker_org_secret" "test_secret" {
	org_id = %d
	name = "%s"
	value = "test123New"
	events = ["%s"]
}
`, org.ID, newName, woodpecker.EventPush),
					ConfigPlanChecks: resource.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{
							plancheck.ExpectResourceAction("woodpecker_org_secret.test_secret", plancheck.ResourceActionReplace),
						},
					},
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("woodpecker_org_secret.test_secret", "id"),
						resource.TestCheckResourceAttr(
							"woodpecker_org_secret.test_secret",
							"org_id",
							strconv.FormatInt(org.ID, 10),
						),
						resource.TestCheckResourceAttr("woodpecker_org_secret.test_secret", "name", newName),
						resource.TestCheckResourceAttr("woodpecker_org_secret.test_secret", "value", "test123New"),
						resource.TestCheckTypeSetElemAttr("woodpecker_org_secret.test_secret", "events.*", woodpecker.EventPush),
					),
				},
				{ // replace secret (new repo id)
					Config: fmt.Sprintf(`
resource "woodpecker_org_secret" "test_secret" {
	org_id = %d
	name = "%s"
	value = "test123New"
	events = ["%s"]
}
`, newOrg.ID, newName, woodpecker.EventPush),
					ConfigPlanChecks: resource.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{
							plancheck.ExpectResourceAction("woodpecker_org_secret.test_secret", plancheck.ResourceActionReplace),
						},
					},
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet("woodpecker_org_secret.test_secret", "id"),
						resource.TestCheckResourceAttr(
							"woodpecker_org_secret.test_secret",
							"org_id",
							strconv.FormatInt(newOrg.ID, 10),
						),
						resource.TestCheckResourceAttr("woodpecker_org_secret.test_secret", "name", newName),
						resource.TestCheckResourceAttr("woodpecker_org_secret.test_secret", "value", "test123New"),
						resource.TestCheckTypeSetElemAttr("woodpecker_org_secret.test_secret", "events.*", woodpecker.EventPush),
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
resource "woodpecker_org_secret" "test_secret" {
	org_id = 123
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

func checkOrgSecretResourceDestroy(m map[int64][]string) func(state *terraform.State) error {
	return func(_ *terraform.State) error {
		for orgID, names := range m {
			secrets, err := woodpeckerClient.OrgSecretList(orgID, woodpecker.SecretListOptions{})
			if err != nil {
				return fmt.Errorf("couldn't list secrets: %w", err)
			}

			if slices.ContainsFunc(secrets, func(secret *woodpecker.Secret) bool {
				return slices.Contains(names, secret.Name)
			}) {
				return errors.New("at least one of the created secrets isn't deleted")
			}
		}

		return nil
	}
}
