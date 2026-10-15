package acceptance_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

// reporterVariables returns the variables of a reporter testdata config, merged with the test's own.
func reporterVariables(domainKey, key string, own config.Variables) config.Variables {
	variables := domainVariables(domainKey, own)
	variables["key"] = config.StringVariable(key)
	return variables
}

func TestReporterResource_crud(t *testing.T) {
	domainKey := uniqueKey("tf-acc-rep")
	key := uniqueKey("audit")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheck(t) },
		ProtoV6ProviderFactories: testProviders(),
		CheckDestroy:             checkDomainsDestroyed,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: reporterVariables(domainKey, key, config.Variables{"enabled": config.BoolVariable(true)}),
				Check:           resource.TestCheckResourceAttr("am_reporter.test", "enabled", "true"),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: reporterVariables(domainKey, key, config.Variables{"enabled": config.BoolVariable(false)}),
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

func TestReporterResource_immutableFields(t *testing.T) {
	requiresLiveAM(t)
	domainKey := uniqueKey("tf-acc-rep")
	key := uniqueKey("audit")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheck(t) },
		ProtoV6ProviderFactories: testProviders(),
		CheckDestroy:             checkDomainsDestroyed,
		Steps: immutableFieldSteps("am_reporter.test", fmt.Sprintf("/domains/%s/reporters/%s", domainKey, key),
			reporterVariables(domainKey, key, nil), "reporter-am-file", "reporter-am-kafka"),
	})
}
