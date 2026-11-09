package acceptance_test

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

// applicationVariables returns the variables of an application testdata config, merged with the test's own.
func applicationVariables(domainKey, key string, own config.Variables) config.Variables {
	variables := domainVariables(domainKey, own)
	variables["key"] = config.StringVariable(key)
	return variables
}

func applicationPath(domainKey, key string) string {
	return fmt.Sprintf("/domains/%s/applications/%s", domainKey, key)
}

// The full example of AM-7512: certificate, identity provider, OAuth settings and metadata.
func TestApplicationResource_crud(t *testing.T) {
	domainKey := uniqueKey("tf-acc-app")
	key := uniqueKey("mobile")
	path := applicationPath(domainKey, key)
	keystore, err := os.ReadFile("testdata/server.jks.b64")
	if err != nil {
		t.Fatal(err)
	}
	created := applicationVariables(domainKey, key, config.Variables{
		"name":         config.StringVariable("Mobile App"),
		"redirect_uri": config.StringVariable("com.example.app://callback"),
		"owner":        config.StringVariable("payments-team"),
		"keystore":     config.StringVariable(strings.TrimSpace(string(keystore))),
	})
	updated := with(with(with(created,
		"name", config.StringVariable("Mobile App renamed")),
		"redirect_uri", config.StringVariable("com.example.app://callback2")),
		"owner", config.StringVariable("identity-team"))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheck(t) },
		ProtoV6ProviderFactories: testProviders(),
		CheckDestroy:             checkDomainsDestroyed,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: created,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("am_application.test", "key", key),
					resource.TestCheckResourceAttr("am_application.test", "certificate", "signing"),
					resource.TestCheckResourceAttr("am_application.test", "identity_providers.0.identity", "users"),
					resource.TestCheckResourceAttrSet("am_application.test", "client_id"),
					checkAM(path, "name", "Mobile App"),
				),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: updated,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("am_application.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					checkAM(path, "name", "Mobile App renamed"),
					checkAMNested(path, "com.example.app://callback2", "settings", "oauth", "redirectUris", "0"),
					checkAMNested(path, "identity-team", "metadata", "owner"),
				),
			},
			importStep("am_application.test", "key", updated, "domain_key", "environment_id", "key", "organization_id"),
		},
	})
}

// A minimal configuration converges, and AM's generated client_id is in state and in the output.
func TestApplicationResource_minimal(t *testing.T) {
	domainKey := uniqueKey("tf-acc-app")
	key := uniqueKey("minimal")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheck(t) },
		ProtoV6ProviderFactories: testProviders(),
		CheckDestroy:             checkDomainsDestroyed,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: applicationVariables(domainKey, key, nil),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("am_application.test", "client_id"),
					resource.TestCheckResourceAttr("am_application.test", "certificate", "created-later"),
					checkOutputMatchesAttribute("client_id", "am_application.test", "client_id"),
					checkAMMatchesState(applicationPath(domainKey, key), "clientId", "am_application.test", "client_id"),
				),
			},
		},
	})
}

// Changing client_id fails the plan instead of replacing the application; dropping it keeps the value.
func TestApplicationResource_clientIdImmutable(t *testing.T) {
	domainKey := uniqueKey("tf-acc-app")
	key := uniqueKey("client-id")
	clientID := uniqueKey("tf-acc-client")
	path := applicationPath(domainKey, key)
	variables := applicationVariables(domainKey, key, config.Variables{"client_id": config.StringVariable(clientID)})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheck(t) },
		ProtoV6ProviderFactories: testProviders(),
		CheckDestroy:             checkDomainsDestroyed,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: variables,
				Check:           checkAM(path, "clientId", clientID),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: with(variables, "client_id", config.StringVariable(clientID+"-changed")),
				ExpectError:     regexp.MustCompile(`client_id cannot be changed once set`),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: applicationVariables(domainKey, key, nil),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("am_application.test", "client_id", clientID),
					checkAM(path, "clientId", clientID),
				),
			},
		},
	})
}

// type updates in place: the OAuth settings AM derives from it follow the new type unless declared.
func TestApplicationResource_typeChange(t *testing.T) {
	domainKey := uniqueKey("tf-acc-app")
	key := uniqueKey("type")
	path := applicationPath(domainKey, key)
	variables := applicationVariables(domainKey, key, config.Variables{"type": config.StringVariable("native")})
	web := with(variables, "type", config.StringVariable("web"))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheck(t) },
		ProtoV6ProviderFactories: testProviders(),
		CheckDestroy:             checkDomainsDestroyed,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: variables,
				// The mock does not derive OAuth settings from the type.
				Check: onLiveAM(checkAMNested(path, "none", "settings", "oauth", "tokenEndpointAuthMethod")),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: web,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("am_application.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					checkAM(path, "type", "web"),
					onLiveAM(checkAMNested(path, "client_secret_basic", "settings", "oauth", "tokenEndpointAuthMethod")),
					onLiveAM(checkAMNested(path, "password", "settings", "oauth", "grantTypes", "1")),
					onLiveAM(resource.TestCheckResourceAttr("am_application.test", "settings.oauth.token_endpoint_auth_method", "client_secret_basic")),
				),
			},
			{
				// Declared derived settings are kept across a type change. Both are declared: AM ignores
				// tokenEndpointAuthMethod when grantTypes is absent, whatever the type.
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: with(with(with(web,
					"type", config.StringVariable("service")),
					"token_endpoint_auth_method", config.StringVariable("client_secret_post")),
					"grant_types", config.ListVariable(config.StringVariable("client_credentials"))),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkAM(path, "type", "service"),
					checkAMNested(path, "client_secret_post", "settings", "oauth", "tokenEndpointAuthMethod"),
					checkAMNested(path, "client_credentials", "settings", "oauth", "grantTypes", "0"),
				),
			},
		},
	})
}

// An omitted login block stays inherited with no diff, a declared one overrides it, and
// inherited = true goes back to the domain's; removing the block afterwards plans nothing.
func TestApplicationResource_inheritedBlocks(t *testing.T) {
	domainKey := uniqueKey("tf-acc-app")
	key := uniqueKey("inherited")
	path := applicationPath(domainKey, key)
	variables := applicationVariables(domainKey, key, nil)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheck(t) },
		ProtoV6ProviderFactories: testProviders(),
		CheckDestroy:             checkDomainsDestroyed,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: variables,
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: variables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("am_application.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: checkAMNested(path, false, "settings", "login", "inherited"),
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: variables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("am_application.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: checkAMNested(path, true, "settings", "login", "inherited"),
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: variables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: checkAMNested(path, true, "settings", "login", "inherited"),
			},
		},
	})
}

// AM rejects a scope that does not exist in the domain; the mock does not check scopes.
func TestApplicationResource_unknownScope(t *testing.T) {
	requiresLiveAM(t)
	domainKey := uniqueKey("tf-acc-app")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheck(t) },
		ProtoV6ProviderFactories: testProviders(),
		CheckDestroy:             checkDomainsDestroyed,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: applicationVariables(domainKey, uniqueKey("scope"), nil),
				ExpectError:     regexp.MustCompile(`tf-acc-unknown-scope`),
			},
		},
	})
}

// The provider retries an application PUT answered with 409 Conflict.
func TestApplicationResource_retriesConflict(t *testing.T) {
	serverURL, conflicts := conflictingMock(t, http.MethodPut, "/applications")
	domainKey := uniqueKey("tf-acc-app")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t) },
		Steps: []resource.TestStep{
			{
				// The config's provider block points at the conflicting mock, so providers are set per step.
				ProtoV6ProviderFactories: testProviders(),
				ConfigDirectory:          config.TestNameDirectory(),
				ConfigVariables:          applicationVariables(domainKey, uniqueKey("retry"), config.Variables{"server_url": config.StringVariable(serverURL)}),
				Check:                    checkConflictsAnswered(conflicts, 1),
			},
		},
	})
}

// Moving an application to another domain replaces it: AM holds it in the new domain only.
func TestApplicationResource_domainChange(t *testing.T) {
	firstDomain, secondDomain := uniqueKey("tf-acc-app"), uniqueKey("tf-acc-app")
	key := uniqueKey("moved")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { preCheck(t) },
		ProtoV6ProviderFactories: testProviders(),
		CheckDestroy:             checkDomainsDestroyed,
		Steps:                    domainChangeSteps("am_application.test", firstDomain, secondDomain, "/applications/"+key, applicationVariables(firstDomain, key, nil)),
	})
}
