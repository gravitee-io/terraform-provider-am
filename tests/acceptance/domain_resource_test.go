package acceptance_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestDomainResource_crud(t *testing.T) {
	key := uniqueKey("tf-acc-domain")
	renamedKey := uniqueKey("tf-acc-domain")
	name := config.StringVariable("Terraform acceptance")
	renamed := config.StringVariable("Terraform acceptance renamed")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheck(t) },
		ProtoV6ProviderFactories: testProviders(),
		CheckDestroy:             checkDomainsDestroyed,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: domainVariables(key, config.Variables{"name": name}),
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
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: domainVariables(key, config.Variables{"name": renamed}),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("am_domain.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.TestCheckResourceAttr("am_domain.test", "name", "Terraform acceptance renamed"),
			},
			{
				ConfigDirectory:                      config.TestNameDirectory(),
				ConfigVariables:                      domainVariables(key, config.Variables{"name": renamed}),
				ResourceName:                         "am_domain.test",
				ImportState:                          true,
				ImportStateIdFunc:                    importStateIDFunc("am_domain.test", "environment_id", "key", "organization_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key",
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: domainVariables(renamedKey, config.Variables{"name": renamed}),
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

// domainChangeSteps creates the resource at address in firstDomain, then moves it to secondDomain:
// the plan replaces it, and AM then holds it at resourcePath in secondDomain only.
// The test's config declares both domains and picks the resource's from the "in_second_domain" variable.
func domainChangeSteps(address, firstDomain, secondDomain, resourcePath string, variables config.Variables) []resource.TestStep {
	variables = with(variables, "second_domain_key", config.StringVariable(secondDomain))
	return []resource.TestStep{
		{
			ConfigDirectory: config.TestNameDirectory(),
			ConfigVariables: with(variables, "in_second_domain", config.BoolVariable(false)),
		},
		{
			ConfigDirectory: config.TestNameDirectory(),
			ConfigVariables: with(variables, "in_second_domain", config.BoolVariable(true)),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(address, plancheck.ResourceActionReplace),
				},
			},
			Check: checkAMMoved("/domains/"+firstDomain+resourcePath, "/domains/"+secondDomain+resourcePath),
		},
	}
}

// checkAMMoved fails unless the Automation API returns to and no longer returns from.
func checkAMMoved(from, to string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		if _, err := automationGet(to); err != nil {
			return err
		}
		if _, err := automationGet(from); err == nil {
			return fmt.Errorf("%s still exists", from)
		}
		return nil
	}
}
