package acceptance_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func certificateConfig(t *testing.T, domainKey, key, name string) string {
	keystore, err := os.ReadFile("testdata/server.jks.b64")
	if err != nil {
		t.Fatal(err)
	}
	return domainConfig(domainKey, "Terraform acceptance") + fmt.Sprintf(`
resource "am_certificate" "test" {
  domain_key = am_domain.test.key
  key        = %[1]q
  name       = %[2]q
  type       = "javakeystore-am-certificate"
  configuration = jsonencode({
    jks = jsonencode({
      name    = "server.jks"
      type    = ""
      size    = 2237
      content = %[3]q
    })
    storepass = "letmein"
    alias     = "mytestkey"
    keypass   = "changeme"
  })
}
`, key, name, strings.TrimSpace(string(keystore)))
}

// Converges although AM masks the keystore file and its passwords.
func TestCertificate_MaskedSecrets(t *testing.T) {
	domainKey := uniqueKey("tf-acc-cert")
	key := uniqueKey("signing")
	path := fmt.Sprintf("/domains/%s/certificates/%s", domainKey, key)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheck(t) },
		ProtoV6ProviderFactories: testProviders(),
		CheckDestroy:             checkDomainsDestroyed,
		Steps: []resource.TestStep{
			{
				Config: certificateConfig(t, domainKey, key, "Terraform acceptance"),
				Check:  checkKeystoreMasked(path),
			},
			{
				Config: certificateConfig(t, domainKey, key, "Terraform acceptance renamed"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("am_certificate.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("am_certificate.test", "name", "Terraform acceptance renamed"),
					checkKeystoreMasked(path),
				),
			},
		},
	})
}

func TestCertificate_ImmutableFields(t *testing.T) {
	domainKey := uniqueKey("tf-acc-cert")
	key := uniqueKey("signing")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheck(t) },
		ProtoV6ProviderFactories: testProviders(),
		CheckDestroy:             checkDomainsDestroyed,
		Steps: immutableFieldSteps("am_certificate.test", fmt.Sprintf("/domains/%s/certificates/%s", domainKey, key),
			certificateConfig(t, domainKey, key, "Terraform acceptance"), "javakeystore-am-certificate", "pkcs12-am-certificate"),
	})
}

func checkKeystoreMasked(path string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		configuration, err := automationConfiguration(path)
		if err != nil {
			return err
		}
		for _, field := range []string{"jks", "storepass", "keypass"} {
			if configuration[field] != mask {
				return fmt.Errorf("AM returned %s as %v instead of %s", field, configuration[field], mask)
			}
		}
		if configuration["alias"] != "mytestkey" {
			return fmt.Errorf("AM returned alias as %v", configuration["alias"])
		}
		return nil
	}
}
