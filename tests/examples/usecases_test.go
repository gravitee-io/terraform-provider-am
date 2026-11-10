package examples_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const rotationUseCase = "application-service_secret_rotation"

// useCaseChecks holds what a guide promises beyond convergence, by use-case.
var useCaseChecks = map[string]resource.TestCheckFunc{
	"application-public_clients_pkce": resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr("am_application.mobile", "client_id", "acme-mobile-app"),
		onLiveAM(
			resource.TestCheckResourceAttr("am_application.spa", "settings.oauth.force_pkce", "true"),
			resource.TestCheckResourceAttr("am_application.spa", "settings.oauth.token_endpoint_auth_method", "none"),
			resource.TestCheckResourceAttr("am_application.mobile", "settings.oauth.force_pkce", "true"),
			resource.TestCheckResourceAttr("am_application.mobile", "settings.oauth.token_endpoint_auth_method", "none"),
		),
	),
}

// Applies each use-case the guides are built from, checks the next plan is empty, and destroys it.
func TestUseCases(t *testing.T) {
	dirs, err := os.ReadDir(filepath.Join(examplesPath, "use-cases"))
	if err != nil {
		t.Fatal(err)
	}
	for _, dir := range dirs {
		t.Run(dir.Name(), func(t *testing.T) {
			t.Parallel()
			step := applyStep(t, useCaseFiles(t, dir.Name()))
			step.Check = useCaseChecks[dir.Name()]
			resource.Test(t, resource.TestCase{
				PreCheck: func() { preCheck(t) },
				Steps:    []resource.TestStep{step},
			})
		})
	}
}

// Follows the rotation of the service guide: add the next secret next to the current one, then delete
// client-secret.tf, which holds the current one.
func TestSecretRotation(t *testing.T) {
	current := useCaseFiles(t, rotationUseCase)
	next := map[string]string{
		"next-client-secret.tf": filepath.Join(examplesPath, "use-cases", rotationUseCase, "rotation", "next-client-secret.tf"),
	}
	both := with(current, next)
	withoutCurrent := with(both, nil)
	delete(withoutCurrent, "client-secret.tf")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t) },
		Steps: []resource.TestStep{
			applyStep(t, current),
			rotationStep(t, both, onLiveAM(
				resource.TestCheckResourceAttrSet("am_client_secret.billing_2026_11", "expires_at"),
				resource.TestCheckResourceAttrSet("am_client_secret.billing_2027_02", "expires_at"),
			)),
			// Terraform deletes billing_2026_11, which fails the step if AM does not.
			rotationStep(t, withoutCurrent, onLiveAM(
				resource.TestCheckResourceAttrSet("am_client_secret.billing_2027_02", "expires_at"),
			)),
		},
	})
}

func rotationStep(t *testing.T, files map[string]string, check resource.TestCheckFunc) resource.TestStep {
	step := applyStep(t, files)
	step.ConfigVariables["next_client_secret"] = config.StringVariable(acctest.RandString(64))
	step.Check = check
	return step
}

// onLiveAM runs the checks of values the mock does not compute (OAuth defaults, secret expiry) on a live AM only.
func onLiveAM(checks ...resource.TestCheckFunc) resource.TestCheckFunc {
	if os.Getenv(mockTestEnv) == "true" {
		return resource.ComposeAggregateTestCheckFunc()
	}
	return resource.ComposeAggregateTestCheckFunc(checks...)
}

func useCaseFiles(t *testing.T, name string) map[string]string {
	paths, err := filepath.Glob(filepath.Join(examplesPath, "use-cases", name, "*.tf"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatalf("use-case %s holds no .tf file", name)
	}
	files := map[string]string{}
	for _, path := range paths {
		files[filepath.Base(path)] = path
	}
	return files
}
