package acceptance_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// dataPlaneVariables returns the variables of a data plane testdata config; its MongoDB is reachable from AM.
func dataPlaneVariables(id, name string) config.Variables {
	return config.Variables{
		"id":           config.StringVariable(id),
		"name":         config.StringVariable(name),
		"mongodb_host": config.StringVariable(envOr("AM_TEST_MONGODB_HOST", "mongodb")),
		"mongodb_port": config.StringVariable(envOr("AM_TEST_MONGODB_PORT", "27017")),
	}
}

func TestDataPlaneResource_crud(t *testing.T) {
	id := uniqueKey("tf-acc-dp")
	path := "/dataplanes/" + id
	renamed := dataPlaneVariables(id, "Terraform acceptance renamed")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheck(t) },
		ProtoV6ProviderFactories: testProviders(),
		CheckDestroy:             checkDataPlanesDestroyed,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: dataPlaneVariables(id, "Terraform acceptance"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("am_data_plane.test", "id", id),
					resource.TestCheckResourceAttr("am_data_plane.test", "type", "mongodb"),
					resource.TestCheckResourceAttrSet("am_data_plane.test", "created_at"),
					checkAM(path, "name", "Terraform acceptance"),
				),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: renamed,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("am_data_plane.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("am_data_plane.test", "name", "Terraform acceptance renamed"),
					checkAM(path, "name", "Terraform acceptance renamed"),
				),
			},
			importStep("am_data_plane.test", "id", renamed, "environment_id", "id", "organization_id"),
		},
	})
}

// Converges although AM never returns a data plane's configuration.
func TestDataPlaneResource_writeOnlyConfiguration(t *testing.T) {
	requiresLiveAM(t)
	id := uniqueKey("tf-acc-dp")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheck(t) },
		ProtoV6ProviderFactories: testProviders(),
		CheckDestroy:             checkDomainsDestroyed,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: dataPlaneVariables(id, "Terraform acceptance"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("am_data_plane.test", "database", "gravitee-am-"+id),
					resource.TestCheckResourceAttr("am_domain.on_data_plane", "data_plane_id", id),
				),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: dataPlaneVariables(id, "Terraform acceptance renamed"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("am_data_plane.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.TestCheckResourceAttr("am_data_plane.test", "name", "Terraform acceptance renamed"),
			},
		},
	})
}

func checkDataPlanesDestroyed(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "am_data_plane" {
			continue
		}
		if _, err := automationGet("/dataplanes/" + rs.Primary.Attributes["id"]); err == nil {
			return fmt.Errorf("data plane %s still exists", rs.Primary.Attributes["id"])
		}
	}
	return nil
}
