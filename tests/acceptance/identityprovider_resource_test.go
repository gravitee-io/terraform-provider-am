package acceptance_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const inlineUserPassword = "Tf-Acc-P@ssw0rd"

// identityProviderVariables returns the variables of an identity provider testdata config, merged with the test's own.
func identityProviderVariables(domainKey, key string, own config.Variables) config.Variables {
	variables := domainVariables(domainKey, own)
	variables["key"] = config.StringVariable(key)
	variables["password"] = config.StringVariable(inlineUserPassword)
	return variables
}

// Converges although AM masks the password in its responses.
func TestIdentityProviderResource_maskedSecret(t *testing.T) {
	requiresLiveAM(t)
	domainKey := uniqueKey("tf-acc-idp")
	key := uniqueKey("users")
	path := fmt.Sprintf("/domains/%s/identities/%s", domainKey, key)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheck(t) },
		ProtoV6ProviderFactories: testProviders(),
		CheckDestroy:             checkDomainsDestroyed,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: identityProviderVariables(domainKey, key, config.Variables{"firstname": config.StringVariable("Alice")}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("am_identity_provider.test", "configuration", containsPassword),
					checkInlinePasswordMasked(path),
				),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: identityProviderVariables(domainKey, key, config.Variables{"firstname": config.StringVariable("Alicia")}),
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
				ConfigDirectory:                      config.TestNameDirectory(),
				ConfigVariables:                      identityProviderVariables(domainKey, key, config.Variables{"firstname": config.StringVariable("Alicia")}),
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
// Step 1 creates the domain only, step 2 imports the identity provider created in between.
func TestIdentityProviderResource_importMaskedSecret(t *testing.T) {
	requiresLiveAM(t)
	domainKey := uniqueKey("tf-acc-idp")
	key := uniqueKey("users")
	importVariables := identityProviderVariables(domainKey, key, config.Variables{
		"organization_id": config.StringVariable(envOr("AM_ORG_ID", "DEFAULT")),
		"environment_id":  config.StringVariable(envOr("AM_ENV_ID", "DEFAULT")),
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheck(t) },
		ProtoV6ProviderFactories: testProviders(),
		CheckDestroy:             checkDomainsDestroyed,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: domainVariables(domainKey, nil),
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
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: importVariables,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("am_identity_provider.test", "configuration", containsPassword),
					checkInlinePasswordMasked(fmt.Sprintf("/domains/%s/identities/%s", domainKey, key)),
				),
			},
		},
	})
}

func TestIdentityProviderResource_immutableFields(t *testing.T) {
	requiresLiveAM(t)
	domainKey := uniqueKey("tf-acc-idp")
	key := uniqueKey("users")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheck(t) },
		ProtoV6ProviderFactories: testProviders(),
		CheckDestroy:             checkDomainsDestroyed,
		Steps: immutableFieldSteps("am_identity_provider.test", fmt.Sprintf("/domains/%s/identities/%s", domainKey, key),
			identityProviderVariables(domainKey, key, nil), "inline-am-idp", "mongo-am-idp"),
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
