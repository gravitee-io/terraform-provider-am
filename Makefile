# Overridable environment variables consumed by the provider at runtime:
#   AM_SERVER_URL  - AM Automation API base URL
#   AM_SA_TOKEN    - bearer (service-account) token
#   AM_ORG_ID      - organization id (defaults to "DEFAULT")
#   AM_ENV_ID      - environment id  (defaults to "DEFAULT")

AM_OAS_BRANCH ?= master

.PHONY: sync-oas
sync-oas: ## Download the AM Automation OAS from upstream into automation-api-oas.yaml
	curl -fsSL "https://raw.githubusercontent.com/gravitee-io/gravitee-access-management/refs/heads/$(AM_OAS_BRANCH)/docs/automation/openapi.yaml" -o automation-api-oas.yaml

.PHONY: speakeasy
speakeasy: ## Generate the provider from automation-api-oas.yaml
	@rm -f terraform-provider-am
	@mv ~/.terraformrc ~/.terraformrc.keep 2>/dev/null || true
	@terraform fmt -recursive > /dev/null 2>&1 || true
	speakeasy run --output console --skip-versioning --skip-upload-spec
	@mv ~/.terraformrc.keep ~/.terraformrc 2>/dev/null || true
	@go build

SCHEMA       := automation-api-oas.yaml
MERGED       := .speakeasy/output/computed.json
OVERLAY_FILES := \
    .speakeasy/overlays/common/oas.yaml \
    .speakeasy/overlays/common/param.yaml \
    .speakeasy/overlays/common/schema.yaml \
    .speakeasy/overlays/domain.yaml \
    .speakeasy/overlays/certificate.yaml \
    .speakeasy/overlays/identities.yaml \
    .speakeasy/overlays/reporter.yaml

.PHONY: merge-overlays
merge-overlays: ## Apply all overlays to the raw OAS, write to .speakeasy/output/computed.json
	@mkdir -p $(dir $(MERGED))
	@cp $(SCHEMA) $(MERGED)
	@for o in $(OVERLAY_FILES); do \
		speakeasy overlay apply --schema $(MERGED) --overlay $$o --strict --out $(MERGED).tmp || exit 1; \
		mv $(MERGED).tmp $(MERGED); \
	done

.PHONY: lint
lint: merge-overlays ## Validate overlays + lint the merged OAS + check Terraform fmt
	@for o in $(OVERLAY_FILES); do \
		speakeasy overlay validate --overlay $$o || exit 1; \
	done
	@speakeasy lint openapi --schema $(MERGED) --max-validation-errors 0 --max-validation-warnings 0 --non-interactive
	@terraform fmt -recursive -check || (echo "Error: Above terraform files are not properly formatted. Please run 'terraform fmt -recursive' to fix formatting issues" && exit 1)

.PHONY: unit-tests
unit-tests: ## Run unit tests (no AM instance required)
	@go test -count=1 ./internal/...

.PHONY: acceptance-tests
acceptance-tests: ## Run acceptance tests against a live AM (not wired up yet)
	@echo "TODO: acceptance tests require a running AM Automation API."
	@echo "      Wire up tests/acceptance/ + a Kind cluster harness, then"
	@echo "      replace this target with the real go test invocation."
	@exit 1

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
