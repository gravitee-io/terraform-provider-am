package acceptance_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// certificateVariables returns the variables of a certificate testdata config, merged with the test's own.
func certificateVariables(t *testing.T, domainKey, key string, own config.Variables) config.Variables {
	keystore, err := os.ReadFile("testdata/server.jks.b64")
	if err != nil {
		t.Fatal(err)
	}
	variables := domainVariables(domainKey, own)
	variables["key"] = config.StringVariable(key)
	variables["keystore"] = config.StringVariable(strings.TrimSpace(string(keystore)))
	return variables
}

func TestCertificateResource_crud(t *testing.T) {
	domainKey := uniqueKey("tf-acc-cert")
	key := uniqueKey("signing")
	path := fmt.Sprintf("/domains/%s/certificates/%s", domainKey, key)
	renamed := certificateVariables(t, domainKey, key, config.Variables{"name": config.StringVariable("Terraform acceptance renamed")})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheck(t) },
		ProtoV6ProviderFactories: testProviders(),
		CheckDestroy:             checkDomainsDestroyed,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: certificateVariables(t, domainKey, key, config.Variables{"name": config.StringVariable("Terraform acceptance")}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("am_certificate.test", "key", key),
					resource.TestCheckResourceAttr("am_certificate.test", "type", "javakeystore-am-certificate"),
					resource.TestCheckResourceAttrSet("am_certificate.test", "created_at"),
					checkAM(path, "name", "Terraform acceptance"),
				),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: renamed,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("am_certificate.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("am_certificate.test", "name", "Terraform acceptance renamed"),
					checkAM(path, "name", "Terraform acceptance renamed"),
				),
			},
			importStep("am_certificate.test", "key", renamed, "domain_key", "environment_id", "key", "organization_id"),
		},
	})
}

// Converges although AM masks the keystore file and its passwords.
func TestCertificateResource_maskedSecrets(t *testing.T) {
	requiresLiveAM(t)
	domainKey := uniqueKey("tf-acc-cert")
	key := uniqueKey("signing")
	path := fmt.Sprintf("/domains/%s/certificates/%s", domainKey, key)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheck(t) },
		ProtoV6ProviderFactories: testProviders(),
		CheckDestroy:             checkDomainsDestroyed,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: certificateVariables(t, domainKey, key, config.Variables{"name": config.StringVariable("Terraform acceptance")}),
				Check:           checkKeystoreMasked(path),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: certificateVariables(t, domainKey, key, config.Variables{"name": config.StringVariable("Terraform acceptance renamed")}),
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

func TestCertificateResource_immutableFields(t *testing.T) {
	requiresLiveAM(t)
	domainKey := uniqueKey("tf-acc-cert")
	key := uniqueKey("signing")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheck(t) },
		ProtoV6ProviderFactories: testProviders(),
		CheckDestroy:             checkDomainsDestroyed,
		Steps: immutableFieldSteps("am_certificate.test", fmt.Sprintf("/domains/%s/certificates/%s", domainKey, key),
			certificateVariables(t, domainKey, key, nil), "javakeystore-am-certificate", "pkcs12-am-certificate"),
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
