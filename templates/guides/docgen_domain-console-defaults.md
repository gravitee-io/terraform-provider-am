---
page_title: "Domain with the console defaults"
subcategory: "Domains"
---

# Domain with the console defaults

Creating a domain in the AM console also creates its default certificate, identity provider and reporter. A
domain created by Terraform gets none of them. This guide declares them, built from AM's defaults.

## Before you start

* The provider reaches the AM Automation API with `AM_SERVER_URL` and a service-account token in `AM_SA_TOKEN`.

## Configuration

```terraform
# A domain with what the AM console creates along with a new domain: a certificate, an identity provider
# holding the users, and a reporter feeding the console's audit log and dashboard. Through Terraform a new domain
# gets none. Any certificate or identity provider can replace the system ones.

resource "am_domain" "partners" {
  key           = "acme-partners"
  name          = "ACME partners"
  description   = "Users of the ACME partner portal"
  path          = "/acme-partners"
  data_plane_id = "default"

  # The login page offers registration; new users are created in the system identity provider below.
  login_settings = {
    inherited        = false
    register_enabled = true
  }
  account_settings = {
    default_identity_provider_for_registration = "default"
  }
}

# system = true builds each of them from AM's domain defaults (the domains.*.default settings of AM):
# only the key is declared.
resource "am_certificate" "default" {
  domain_key = am_domain.partners.key
  key        = "default"
  system     = true
}

resource "am_identity_provider" "default" {
  domain_key = am_domain.partners.key
  key        = "default"
  system     = true
}

resource "am_reporter" "default" {
  domain_key = am_domain.partners.key
  key        = "default"
  system     = true
}

# An application signs its tokens with the certificate it names, and lists the identity providers of its login page.
resource "am_application" "portal" {
  domain_key  = am_domain.partners.key
  key         = "partner-portal"
  name        = "Partner portal"
  type        = "web"
  certificate = am_certificate.default.key

  identity_providers = [
    { identity = am_identity_provider.default.key }
  ]

  settings = {
    oauth = {
      redirect_uris = ["https://partners.example.com/callback"]
    }
  }
}

```

## How it differs from the console

* The certificate and the identity provider can be of any type: the system ones are only the console's
  defaults. A domain needs at least one certificate to sign tokens with, and one identity provider to sign in with
  or register in.
* The console's audit log and dashboard read a reporter that stores the events in a database, such as the system
  reporter: without one, they stay empty. A reporter that only sends the events elsewhere (Kafka, file, TCP) does
  not fill them.
* An application does not pick the domain's certificate or identity providers by itself: name them in its
  `certificate` (required) and `identity_providers`, as `am_application.portal` does.
* `system = true` builds the resource from AM's domain defaults, the `domains.certificates.default`,
  `domains.identities.default` and `domains.reporters.default` settings of the `gravitee.yml` of the AM management
  API. Only `key` is
  read: `name`, `type` and `configuration` are ignored. A domain has at most one system resource of each kind.
* `system` and `type` cannot change once created: AM rejects the apply. Remove the resource, apply, then
  declare the new one.
* Registration needs both `login_settings.register_enabled`, which shows the sign-up link, and
  `default_identity_provider_for_registration`, which tells AM where to create the users. The login settings are
  inherited from the organization, where registration is off, unless `inherited = false`.
* `default_identity_provider_for_registration` names the identity provider by its key, as a plain string.
  `am_identity_provider.default.key` would be a dependency cycle, since the identity provider needs the
  domain's key. AM does not check the reference: it resolves whenever an identity provider with that key exists,
  and stays when that identity provider is deleted.
