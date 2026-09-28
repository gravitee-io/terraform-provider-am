package acceptance_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func dataPlaneConfig(id, name string) string {
	return fmt.Sprintf(`
resource "am_data_plane" "test" {
  id   = %[1]q
  name = %[2]q
  type = "mongodb"
  configuration = jsonencode({
    mongodb = {
      dbname = "gravitee-am-%[1]s"
      host   = %[3]q
      port   = %[4]s
    }
  })
}

resource "am_domain" "on_data_plane" {
  key           = "%[1]s-domain"
  name          = "Terraform acceptance on a data plane"
  path          = "/%[1]s-domain"
  data_plane_id = am_data_plane.test.id
}
`, id, name, envOr("AM_TEST_MONGODB_HOST", "mongodb"), envOr("AM_TEST_MONGODB_PORT", "27017"))
}

// Converges although AM never returns a data plane's configuration.
func TestDataPlane(t *testing.T) {
	id := uniqueKey("tf-acc-dp")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheck(t) },
		ProtoV6ProviderFactories: testProviders(),
		CheckDestroy:             checkDomainsDestroyed,
		Steps: []resource.TestStep{
			{
				Config: dataPlaneConfig(id, "Terraform acceptance"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("am_data_plane.test", "database", "gravitee-am-"+id),
					resource.TestCheckResourceAttr("am_domain.on_data_plane", "data_plane_id", id),
				),
			},
			{
				Config: dataPlaneConfig(id, "Terraform acceptance renamed"),
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
