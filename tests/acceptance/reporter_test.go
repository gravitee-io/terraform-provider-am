package acceptance_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func reporterConfig(domainKey, key string, enabled bool) string {
	return domainConfig(domainKey, "Terraform acceptance") + fmt.Sprintf(`
resource "am_reporter" "test" {
  domain_key = am_domain.test.key
  key        = %[1]q
  name       = "Terraform acceptance audit"
  type       = "reporter-am-file"
  enabled    = %[2]t
  configuration = jsonencode({
    filename = %[1]q
  })
}
`, key, enabled)
}

func TestReporter(t *testing.T) {
	domainKey := uniqueKey("tf-acc-rep")
	key := uniqueKey("audit")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheck(t) },
		ProtoV6ProviderFactories: testProviders(),
		CheckDestroy:             checkDomainsDestroyed,
		Steps: []resource.TestStep{
			{
				Config: reporterConfig(domainKey, key, true),
				Check:  resource.TestCheckResourceAttr("am_reporter.test", "enabled", "true"),
			},
			{
				Config: reporterConfig(domainKey, key, false),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("am_reporter.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.TestCheckResourceAttr("am_reporter.test", "enabled", "false"),
			},
		},
	})
}
