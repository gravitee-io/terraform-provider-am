---
page_title: "Sign in with an external OpenID Connect provider"
subcategory: "Identity providers"
---

# Sign in with an external OpenID Connect provider

This guide lets the employees of a company sign in to an internal application with the company's own OpenID
Connect provider: AM is a client of that provider, and the application only talks to AM.

## Before you start

* The provider reaches the AM Automation API with `AM_SERVER_URL` and a service-account token in `AM_SA_TOKEN`.
* AM registered as a client at the company's provider, with the redirect URI
  `<gateway URL>/acme-intranet/login/callback`; its secret in `TF_VAR_corporate_client_secret`.

## Configuration

```terraform
# Employees sign in to an internal application with the company's OpenID Connect provider.

resource "am_domain" "intranet" {
  key           = "acme-intranet"
  name          = "ACME intranet"
  path          = "/acme-intranet"
  data_plane_id = "default"
}

# Signs the tokens, built from AM's default certificate settings (see the domain with the console defaults guide).
resource "am_certificate" "default" {
  domain_key = am_domain.intranet.key
  key        = "default"
  system     = true
}

resource "am_identity_provider" "corporate" {
  domain_key = am_domain.intranet.key
  key        = "corporate-sso"
  name       = "ACME corporate SSO"
  type       = "oauth2-generic-am-idp"

  # AM returns clientSecret as ********: the provider keeps the configured value instead of seeing a change.
  configuration = jsonencode({
    clientId     = "am-intranet"
    clientSecret = var.corporate_client_secret
    wellKnownUri = "https://sso.example.com/.well-known/openid-configuration"
    responseType = "code"
    responseMode = "default"
    scopes       = ["openid", "profile", "email"]
  })

  # Only accounts of these email domains can sign in.
  domain_whitelist = ["example.com"]

  # AM user profile attribute = claim of the corporate provider, or an expression on the claims (profile).
  # Once set, AM keeps only the mapped attributes: map the standard claims too.
  mappers = {
    email       = "email"
    given_name  = "given_name"
    family_name = "family_name"
    department  = "dept"
    cost_center = "{#profile['org']['cost_center']}"
  }
}

resource "am_application" "intranet" {
  domain_key  = am_domain.intranet.key
  key         = "intranet-portal"
  name        = "Intranet portal"
  type        = "web"
  certificate = am_certificate.default.key

  identity_providers = [
    { identity = am_identity_provider.corporate.key }
  ]

  settings = {
    oauth = {
      redirect_uris = ["https://intranet.example.com/callback"]
      scope_settings = [
        { scope = "openid", default_scope = true },
        { scope = "profile", default_scope = true },
        { scope = "email", default_scope = true },
      ]
    }
  }
}

output "intranet_client_id" {
  value = am_application.intranet.client_id
}

# The secret AM authenticates with at the corporate provider.
variable "corporate_client_secret" {
  type      = string
  sensitive = true
}

```

## How it differs from the console

* AM returns the secrets of a plugin configuration as `********`. The provider compares the configuration
  with that masking taken into account and keeps the configured values in the state: a secret changed in the
  console does not show in the plan. Change the variable instead, and apply.
* After an import, the state holds only `********`: the first plan shows the configuration as changed, and the
  apply sends the configured secrets again.
* `configuration` is the JSON of the `oauth2-generic-am-idp` plugin, with the field names of the plugin, not the
  labels of the console form.
* `mappers` maps an AM user profile attribute (the key) to a claim of the provider, or to an expression on its
  claims such as `{#profile['org']['cost_center']}` (the value). Once `mappers` is set, AM keeps only the
  mapped attributes, not the default mapping of the plugin: map `email` and the names too, as here.
* `role_mapper` and `group_mapper` name AM roles and groups by their ID: Terraform does not manage roles and
  groups, so read their IDs in the console.
* `type` cannot change once created: AM rejects the apply.
* The application shows the identity providers listed in its `identity_providers` on its login page.
