package acceptance_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// Tests whose config uses the random provider (ephemeral random_password feeding value_wo) set the
// providers per step: with providers set on the TestCase, plugin-testing does not install random.

// clientSecretVariables returns the variables of a client secret testdata config, merged with the test's own.
func clientSecretVariables(domainKey, applicationKey string, own config.Variables) config.Variables {
	variables := domainVariables(domainKey, own)
	variables["application_key"] = config.StringVariable(applicationKey)
	return variables
}

func clientSecretsPath(domainKey, applicationKey string) string {
	return fmt.Sprintf("/domains/%s/applications/%s/secrets", domainKey, applicationKey)
}

func TestClientSecretResource_crud(t *testing.T) {
	domainKey := uniqueKey("tf-acc-secret")
	applicationKey := uniqueKey("app")
	value := acctest.RandString(64)
	variables := clientSecretVariables(domainKey, applicationKey, config.Variables{"value": config.StringVariable(value)})
	path := clientSecretsPath(domainKey, applicationKey) + "/secret-2026-10"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheck(t) },
		ProtoV6ProviderFactories: testProviders(),
		CheckDestroy:             checkDomainsDestroyed,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: variables,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("am_client_secret.test", "key", "secret-2026-10"),
					resource.TestCheckNoResourceAttr("am_client_secret.test", "value_wo"),
					checkStateHasNo(value),
					// The mock does not compute the expiry from secret_expiration_settings.
					onLiveAM(resource.TestCheckResourceAttrSet("am_client_secret.test", "expires_at")),
					checkAMMatchesState(path, "expiresAt", "am_client_secret.test", "expires_at"),
					checkAMHasNo(path, "value"),
				),
			},
			importStep("am_client_secret.test", "key", variables, "application_key", "domain_key", "environment_id", "key", "organization_id"),
		},
	})
}

// A new value for an existing secret is invisible to Terraform: no diff, no API call.
func TestClientSecretResource_valueOnlyChange(t *testing.T) {
	variables := clientSecretVariables(uniqueKey("tf-acc-secret"), uniqueKey("app"), config.Variables{"length": config.IntegerVariable(64)})

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { preCheck(t) },
		CheckDestroy: checkDomainsDestroyed,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: testProviders(),
				ConfigDirectory:          config.TestNameDirectory(),
				ConfigVariables:          variables,
			},
			{
				ProtoV6ProviderFactories: testProviders(),
				ConfigDirectory:          config.TestNameDirectory(),
				ConfigVariables:          with(variables, "length", config.IntegerVariable(65)),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// The two-apply rotation: add the new secret next to the old one, then remove the old one.
func TestClientSecretResource_rotation(t *testing.T) {
	domainKey := uniqueKey("tf-acc-secret")
	applicationKey := uniqueKey("app")
	variables := clientSecretVariables(domainKey, applicationKey, nil)
	path := clientSecretsPath(domainKey, applicationKey)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { preCheck(t) },
		CheckDestroy: checkDomainsDestroyed,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: testProviders(),
				ConfigDirectory:          config.TestStepDirectory(),
				ConfigVariables:          variables,
				Check:                    checkAMSecretKeys(path, "secret-2026-10"),
			},
			{
				ProtoV6ProviderFactories: testProviders(),
				ConfigDirectory:          config.TestStepDirectory(),
				ConfigVariables:          variables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("am_client_secret.secret_2026_10", plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction("am_client_secret.secret_2026_11", plancheck.ResourceActionCreate),
					},
				},
				Check: checkAMSecretKeys(path, "secret-2026-10", "secret-2026-11"),
			},
			{
				ProtoV6ProviderFactories: testProviders(),
				ConfigDirectory:          config.TestStepDirectory(),
				ConfigVariables:          variables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("am_client_secret.secret_2026_10", plancheck.ResourceActionDestroy),
						plancheck.ExpectResourceAction("am_client_secret.secret_2026_11", plancheck.ResourceActionNoop),
					},
				},
				Check: checkAMSecretKeys(path, "secret-2026-11"),
			},
		},
	})
}

// Ten secrets, the default per-application limit, created in parallel on one application. On a live AM
// their writes can collide and get a 409; the retry itself is proven by the retriesConflict tests.
func TestClientSecretResource_parallel(t *testing.T) {
	domainKey := uniqueKey("tf-acc-secret")
	applicationKey := uniqueKey("app")
	keys := make([]string, 10)
	for i := range keys {
		keys[i] = fmt.Sprintf("secret-%d", i)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { preCheck(t) },
		CheckDestroy: checkDomainsDestroyed,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: testProviders(),
				ConfigDirectory:          config.TestNameDirectory(),
				ConfigVariables:          clientSecretVariables(domainKey, applicationKey, nil),
				Check:                    checkAMSecretKeys(clientSecretsPath(domainKey, applicationKey), keys...),
			},
		},
	})
}

// The provider retries a secret PUT answered with 409 Conflict.
func TestClientSecretResource_retriesConflict(t *testing.T) {
	serverURL, conflicts := conflictingMock(t, http.MethodPut, "/secrets")
	variables := clientSecretVariables(uniqueKey("tf-acc-secret"), uniqueKey("app"), config.Variables{"server_url": config.StringVariable(serverURL)})

	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t) },
		Steps: []resource.TestStep{
			{
				// The config's provider block points at the conflicting mock, so providers are set per step.
				ProtoV6ProviderFactories: testProviders(),
				ConfigDirectory:          config.TestNameDirectory(),
				ConfigVariables:          variables,
				Check:                    checkConflictsAnswered(conflicts, 1),
			},
		},
	})
}

// checkAMSecretKeys fails unless the Automation API lists exactly keys (sorted) as the secrets of path.
func checkAMSecretKeys(path string, keys ...string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		body, err := automationCall(http.MethodGet, path, nil)
		if err != nil {
			return err
		}
		var secrets []struct {
			Key string `json:"key"`
		}
		if err := json.Unmarshal(body, &secrets); err != nil {
			return err
		}
		var got []string
		for _, secret := range secrets {
			got = append(got, secret.Key)
		}
		slices.Sort(got)
		if !slices.Equal(got, keys) {
			return fmt.Errorf("AM lists secrets %v instead of %v", got, keys)
		}
		return nil
	}
}

// checkAMHasNo fails if the Automation API returns field for path.
func checkAMHasNo(path, field string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		body, err := automationGet(path)
		if err != nil {
			return err
		}
		if _, found := body[field]; found {
			return fmt.Errorf("AM returned %s for %s", field, path)
		}
		return nil
	}
}

// checkStateHasNo fails if any attribute of any resource in state holds value.
func checkStateHasNo(value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for address, rs := range s.RootModule().Resources {
			for attribute, held := range rs.Primary.Attributes {
				if strings.Contains(held, value) {
					return fmt.Errorf("state holds the secret value in %s.%s", address, attribute)
				}
			}
		}
		return nil
	}
}

// The provider retries a secret DELETE answered with 409 Conflict.
func TestClientSecretResource_retriesDeleteConflict(t *testing.T) {
	serverURL, conflicts := conflictingMock(t, http.MethodDelete, "/secrets/secret-2026-10")
	variables := clientSecretVariables(uniqueKey("tf-acc-secret"), uniqueKey("app"), config.Variables{"server_url": config.StringVariable(serverURL)})

	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t) },
		// Runs after the destroy, which deletes the secret through the conflicting mock.
		CheckDestroy: checkConflictsAnswered(conflicts, 1),
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: testProviders(),
				// Same config as the PUT retry test: the conflict only differs by method.
				ConfigDirectory: config.StaticDirectory("testdata/TestClientSecretResource_retriesConflict"),
				ConfigVariables: variables,
			},
		},
	})
}

// Moving a secret to another application replaces it: AM holds it under the new application only.
func TestClientSecretResource_applicationChange(t *testing.T) {
	domainKey := uniqueKey("tf-acc-secret")
	firstApplication, secondApplication := uniqueKey("app"), uniqueKey("app")
	variables := clientSecretVariables(domainKey, firstApplication, config.Variables{
		"second_application_key": config.StringVariable(secondApplication),
	})
	secretPath := func(applicationKey string) string {
		return clientSecretsPath(domainKey, applicationKey) + "/secret-2026-10"
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { preCheck(t) },
		CheckDestroy: checkDomainsDestroyed,
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: testProviders(),
				ConfigDirectory:          config.TestNameDirectory(),
				ConfigVariables:          with(variables, "in_second_application", config.BoolVariable(false)),
			},
			{
				ProtoV6ProviderFactories: testProviders(),
				ConfigDirectory:          config.TestNameDirectory(),
				ConfigVariables:          with(variables, "in_second_application", config.BoolVariable(true)),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("am_client_secret.test", plancheck.ResourceActionReplace),
					},
				},
				Check: checkAMMoved(secretPath(firstApplication), secretPath(secondApplication)),
			},
		},
	})
}
