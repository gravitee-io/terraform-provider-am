package acceptance_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestDomain(t *testing.T) {
	key := uniqueKey("tf-acc-domain")
	renamedKey := uniqueKey("tf-acc-domain")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheck(t) },
		ProtoV6ProviderFactories: testProviders(),
		CheckDestroy:             checkDomainsDestroyed,
		Steps: []resource.TestStep{
			{
				Config: domainConfig(key, "Terraform acceptance"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("am_domain.test", "key", key),
					resource.TestCheckResourceAttr("am_domain.test", "name", "Terraform acceptance"),
					func(*terraform.State) error {
						_, err := automationGet("/domains/" + key)
						return err
					},
				),
			},
			{
				Config: domainConfig(key, "Terraform acceptance renamed"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("am_domain.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.TestCheckResourceAttr("am_domain.test", "name", "Terraform acceptance renamed"),
			},
			{
				ResourceName:                         "am_domain.test",
				ImportState:                          true,
				ImportStateIdFunc:                    importStateIDFunc("am_domain.test", "environment_id", "key", "organization_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key",
			},
			{
				Config: domainConfig(renamedKey, "Terraform acceptance renamed"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("am_domain.test", plancheck.ResourceActionReplace),
					},
				},
				Check: func(*terraform.State) error {
					if _, err := automationGet("/domains/" + key); err == nil {
						return fmt.Errorf("domain %s is still present after its key changed", key)
					}
					return nil
				},
			},
		},
	})
}

func checkDomainsDestroyed(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "am_domain" {
			continue
		}
		if _, err := automationGet("/domains/" + rs.Primary.Attributes["key"]); err == nil {
			return fmt.Errorf("domain %s still exists", rs.Primary.Attributes["key"])
		}
	}
	return nil
}
