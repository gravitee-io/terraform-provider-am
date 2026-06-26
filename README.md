# Gravitee - AM Terraform Provider

<picture>
  <source media="(prefers-color-scheme: dark)" srcset=".assets/gravitee-logo-dark.svg">
  <source media="(prefers-color-scheme: light)" srcset=".assets/gravitee-logo-light.svg">
  <img alt="Gravitee.io" width="400">
</picture>

<!-- Start Summary [summary] -->
## Summary

Gravitee: Gravitee Access Management Terraform Provider (alpha)

You can manage with Terraform the following:
* Domains
* Certificates
* Identity Providers
* Reporters

Compatible with Gravitee Access Management 4.12.0.
<!-- End Summary [summary] -->

<!-- Start Table of Contents [toc] -->
## Table of Contents
<!-- $toc-max-depth=2 -->
* [Gravitee - AM Terraform Provider](#gravitee-am-terraform-provider)
  * [Installation](#installation)
  * [Authentication](#authentication)
  * [Available Resources and Data Sources](#available-resources-and-data-sources)

<!-- End Table of Contents [toc] -->

<!-- Start Installation [installation] -->
## Installation

To install this provider, copy and paste this code into your Terraform configuration. Then, run `terraform init`.

```hcl
terraform {
  required_providers {
    am = {
      source  = "gravitee-io/am"
      version = "0.0.1"
    }
  }
}

provider "am" {
  server_url = "..." # Optional - can use AM_SERVER_URL environment variable
}
```
<!-- End Installation [installation] -->

<!-- Start Authentication [security] -->
## Authentication

This provider supports authentication configuration via environment variables and provider configuration.

The configuration precedence is:

- Provider configuration
- Environment variables

Available configuration:

| Provider Attribute | Description |
|---|---|
| `bearer_auth` | Authentication uses a bearer token: a JWT, or an opaque user service-account access token. Every operation is additionally permission-gated against the target organization, environment, and resource; a caller lacking the required permission receives a 403 response. Configurable via environment variable `AM_SA_TOKEN`. |
<!-- End Authentication [security] -->

<!-- Start Available Resources and Data Sources [operations] -->
## Available Resources and Data Sources

### Managed Resources

* [am_certificate](docs/resources/certificate.md)
* [am_domain](docs/resources/domain.md)
* [am_identity_provider](docs/resources/identity_provider.md)
* [am_reporter](docs/resources/reporter.md)

### Data Sources

* [am_certificate](docs/data-sources/certificate.md)
* [am_domain](docs/data-sources/domain.md)
* [am_identity_provider](docs/data-sources/identity_provider.md)
* [am_reporter](docs/data-sources/reporter.md)
<!-- End Available Resources and Data Sources [operations] -->

<!-- No End Testing the provider locally [usage] -->

<!-- Placeholder for Future Speakeasy SDK Sections -->
