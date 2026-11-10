---
page_title: "Domain with the console defaults"
subcategory: "Domains"
---

# Domain with the console defaults

Creating a domain in the AM console also creates its default certificate, identity provider and reporter. A
domain created by Terraform gets none of them. This guide declares them.

## Before you start

* The provider reaches the AM Automation API with `AM_SERVER_URL` and a service-account token in `AM_SA_TOKEN`.

## Configuration

```terraform
# A domain with what the AM console creates along with a new domain: a certificate, an identity provider
# holding the users, and a reporter feeding the console's audit log. Through Terraform a new domain gets none.

resource "am_domain" "partners" {
  key           = "acme-partners"
  name          = "ACME partners"
  description   = "Users of the ACME partner portal"
  path          = "/acme-partners"
  data_plane_id = "default"

  # New users register in the system identity provider below, named by its key.
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

```

## How it differs from the console

* Without these resources, the domain has no certificate for the applications that name none, no identity
  provider where users can register, and an empty audit log in the console, which reads the system reporter.
* `system = true` builds the resource from AM's domain defaults, the `domains.certificates.default`,
  `domains.identities.default` and `domains.reporters.default` settings of the AM management API. Only `key` is
  read: `name`, `type` and `configuration` are ignored. A domain has at most one system resource of each kind.
* `system` and `type` cannot change once created: AM rejects the apply. Remove the resource, apply, then
  declare the new one.
* `default_identity_provider_for_registration` names the identity provider by its key, as a plain string.
  `am_identity_provider.default.key` would be a dependency cycle, since the identity provider needs the
  domain's key. AM does not check the reference: it resolves whenever an identity provider with that key exists,
  and stays when that identity provider is deleted.
