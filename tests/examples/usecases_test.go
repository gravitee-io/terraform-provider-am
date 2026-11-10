package examples_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	secretRotationUseCase      = "application-service_secret_rotation"
	certificateRotationUseCase = "certificate-signing_rotation"
)

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
	// The guide reads hosts and database back to check what AM connects to.
	"data_plane-dedicated_database": onLiveAM(
		resource.TestCheckResourceAttr("am_data_plane.eu", "hosts.#", "1"),
		resource.TestCheckResourceAttr("am_data_plane.eu", "database", "acme-eu"),
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
	current := useCaseFiles(t, secretRotationUseCase)
	next := map[string]string{
		"next-client-secret.tf": rotationFile(secretRotationUseCase, "next-client-secret.tf"),
	}
	both := with(current, next)
	withoutCurrent := with(both, nil)
	delete(withoutCurrent, "client-secret.tf")

	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t) },
		Steps: []resource.TestStep{
			applyStep(t, current),
			checkedStep(t, both, onLiveAM(
				resource.TestCheckResourceAttrSet("am_client_secret.billing_2026_11", "expires_at"),
				resource.TestCheckResourceAttrSet("am_client_secret.billing_2027_02", "expires_at"),
			)),
			// Terraform deletes billing_2026_11, which fails the step if AM does not.
			checkedStep(t, withoutCurrent, onLiveAM(
				resource.TestCheckResourceAttrSet("am_client_secret.billing_2027_02", "expires_at"),
			)),
		},
	})
}

// Follows the rotation of the certificate guide: add the next certificate, switch the domain and the
// application to it, then delete the current one.
func TestCertificateRotation(t *testing.T) {
	current := useCaseFiles(t, certificateRotationUseCase)
	withNext := with(current, map[string]string{
		"signing-2027.tf": rotationFile(certificateRotationUseCase, "signing-2027.tf"),
	})
	switched := with(withNext, map[string]string{
		"current-certificate.tf": rotationFile(certificateRotationUseCase, "current-certificate.tf"),
	})
	withoutCurrent := with(switched, nil)
	delete(withoutCurrent, "signing-2026.tf")

	signedBy := func(key string) resource.TestCheckFunc {
		return resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("am_domain.accounts", "certificate_settings.fallback_certificate", key),
			resource.TestCheckResourceAttr("am_application.account_portal", "certificate", key),
		)
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() { preCheck(t) },
		Steps: []resource.TestStep{
			checkedStep(t, current, signedBy("signing-2026")),
			checkedStep(t, withNext, signedBy("signing-2026")),
			checkedStep(t, switched, signedBy("signing-2027")),
			// Terraform deletes signing_2026, which fails the step if AM does not.
			checkedStep(t, withoutCurrent, signedBy("signing-2027")),
		},
	})
}

func checkedStep(t *testing.T, files map[string]string, check resource.TestCheckFunc) resource.TestStep {
	step := applyStep(t, files)
	step.Check = check
	return step
}

func rotationFile(useCase, name string) string {
	return filepath.Join(examplesPath, "use-cases", useCase, "rotation", name)
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
