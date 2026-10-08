package acceptance_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const inlineUserPassword = "Tf-Acc-P@ssw0rd"

func identityProviderConfig(domainKey, key, firstname string) string {
	return domainConfig(domainKey, "Terraform acceptance") + fmt.Sprintf(`
resource "am_identity_provider" "test" {
  domain_key = am_domain.test.key
  key        = %[1]q
  name       = "Terraform acceptance users"
  type       = "inline-am-idp"
  configuration = jsonencode({
    users = [{
      firstname = %[2]q
      lastname  = "Acceptance"
      username  = "alice"
      password  = %[3]q
    }]
  })
}
`, key, firstname, inlineUserPassword)
}

// Converges although AM masks the password in its responses.
func TestIdentityProvider_MaskedSecret(t *testing.T) {
	domainKey := uniqueKey("tf-acc-idp")
	key := uniqueKey("users")
	path := fmt.Sprintf("/domains/%s/identities/%s", domainKey, key)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheck(t) },
		ProtoV6ProviderFactories: testProviders(),
		CheckDestroy:             checkDomainsDestroyed,
		Steps: []resource.TestStep{
			{
				Config: identityProviderConfig(domainKey, key, "Alice"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("am_identity_provider.test", "configuration", containsPassword),
					checkInlinePasswordMasked(path),
				),
			},
			{
				Config: identityProviderConfig(domainKey, key, "Alicia"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("am_identity_provider.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("am_identity_provider.test", "configuration", containsPassword),
					checkInlinePasswordMasked(path),
				),
			},
			{
				ResourceName:                         "am_identity_provider.test",
				ImportState:                          true,
				ImportStateIdFunc:                    importStateIDFunc("am_identity_provider.test", "domain_key", "environment_id", "key", "organization_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key",
				ImportStateVerifyIgnore:              []string{"configuration"},
			},
		},
	})
}

// Adopts an identity provider created outside Terraform; the first apply sends the configured password.
func TestIdentityProvider_ImportMaskedSecret(t *testing.T) {
	domainKey := uniqueKey("tf-acc-idp")
	key := uniqueKey("users")
	importBlock := fmt.Sprintf(`
import {
  to = am_identity_provider.test
  id = jsonencode({ domain_key = %q, environment_id = %q, key = %q, organization_id = %q })
}
`, domainKey, envOr("AM_ENV_ID", "DEFAULT"), key, envOr("AM_ORG_ID", "DEFAULT"))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheck(t) },
		ProtoV6ProviderFactories: testProviders(),
		CheckDestroy:             checkDomainsDestroyed,
		Steps: []resource.TestStep{
			{
				Config: domainConfig(domainKey, "Terraform acceptance"),
			},
			{
				PreConfig: func() {
					err := automationPut(fmt.Sprintf("/domains/%s/identities", domainKey), map[string]any{
						"key":  key,
						"name": "Terraform acceptance users",
						"type": "inline-am-idp",
						"configuration": fmt.Sprintf(
							`{"users":[{"firstname":"Alice","lastname":"Acceptance","username":"alice","password":%q}]}`,
							inlineUserPassword),
					})
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: identityProviderConfig(domainKey, key, "Alice") + importBlock,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("am_identity_provider.test", "configuration", containsPassword),
					checkInlinePasswordMasked(fmt.Sprintf("/domains/%s/identities/%s", domainKey, key)),
				),
			},
		},
	})
}

func TestIdentityProvider_ImmutableFields(t *testing.T) {
	domainKey := uniqueKey("tf-acc-idp")
	key := uniqueKey("users")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheck(t) },
		ProtoV6ProviderFactories: testProviders(),
		CheckDestroy:             checkDomainsDestroyed,
		Steps: immutableFieldSteps("am_identity_provider.test", fmt.Sprintf("/domains/%s/identities/%s", domainKey, key),
			identityProviderConfig(domainKey, key, "Alice"), "inline-am-idp", "mongo-am-idp"),
	})
}

func containsPassword(configuration string) error {
	if !strings.Contains(configuration, inlineUserPassword) {
		return fmt.Errorf("state holds %s instead of the configured password", configuration)
	}
	return nil
}

func checkInlinePasswordMasked(path string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		configuration, err := automationConfiguration(path)
		if err != nil {
			return err
		}
		users, _ := configuration["users"].([]any)
		if len(users) != 1 {
			return fmt.Errorf("expected one inline user, got %v", configuration["users"])
		}
		if password := users[0].(map[string]any)["password"]; password != mask {
			return fmt.Errorf("AM returned the inline password as %v instead of %s", password, mask)
		}
		return nil
	}
}
