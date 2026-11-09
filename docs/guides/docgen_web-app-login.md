---
page_title: "Web application login"
subcategory: "Applications"
---

# Web application login

This guide creates everything a server-side web application needs to sign its users in with OpenID Connect:
a domain, a signing certificate, an identity provider holding the users, the application, and its client secret.

## Before you start

* The provider reaches the AM Automation API with `AM_SERVER_URL` and a service-account token in `AM_SA_TOKEN`.
* Terraform 1.11 or later: the client secret value is a write-only attribute, never stored in the state.
* A Java keystore holding the RSA key that signs the tokens, next to the configuration:

```shell
keytool -genkeypair -alias signing -keyalg RSA -keysize 2048 -validity 365 -dname CN=acme-shop \
  -storetype JKS -keystore keystore.jks -storepass:env TF_VAR_keystore_password -keypass:env TF_VAR_key_password
```

## Provider

```terraform
terraform {
  # Write-only attributes, such as the value_wo of am_client_secret, need Terraform 1.11 or later.
  required_version = ">= 1.11"
  required_providers {
    am = {
      source = "gravitee-io/am"
    }
  }
}

provider "am" {
  # Can be set/overridden with the AM_SERVER_URL environment variable
  #server_url = "http://localhost:8093/automation"

  # Can be set/overridden with the AM_SA_TOKEN environment variable
  #bearer_auth = "xxx" # a JWT or an opaque service-account access token

  # organization_id and environment_id default to "DEFAULT".
  # Can be set/overridden with the AM_ORG_ID / AM_ENV_ID environment variables.
  #organization_id = "DEFAULT"
  #environment_id  = "DEFAULT"
}

```

## Configuration

```terraform
# A web application signing in the users of its domain with OpenID Connect (authorization code flow).

resource "am_domain" "shop" {
  key           = "acme-shop"
  name          = "ACME shop"
  description   = "Customers of the ACME online shop"
  path          = "/acme-shop"
  data_plane_id = "default"

  # Overrides the organization's login settings: users can register and reset their password.
  login_settings = {
    inherited               = false
    register_enabled        = true
    forgot_password_enabled = true
  }
}

# Signs the ID and access tokens. keystore.jks holds an RSA key pair under the alias "signing".
resource "am_certificate" "signing" {
  domain_key = am_domain.shop.key
  key        = "shop-signing"
  name       = "Shop signing key"
  type       = "javakeystore-am-certificate"
  configuration = jsonencode({
    jks = jsonencode({
      name    = "keystore.jks"
      content = filebase64("${path.module}/keystore.jks")
    })
    storepass = var.keystore_password
    alias     = "signing"
    keypass   = var.key_password
  })
}

# Users declared inline, enough to try the login page. Use an LDAP, database or social provider for real users.
resource "am_identity_provider" "customers" {
  domain_key = am_domain.shop.key
  key        = "customers"
  name       = "Customers"
  type       = "inline-am-idp"
  configuration = jsonencode({
    users = [{
      firstname = "Alice"
      lastname  = "Example"
      username  = "alice"
      password  = var.alice_password
    }]
  })
}

resource "am_application" "web" {
  domain_key  = am_domain.shop.key
  key         = "shop-web"
  name        = "ACME shop website"
  type        = "web"
  certificate = am_certificate.signing.key

  identity_providers = [
    { identity = am_identity_provider.customers.key }
  ]

  settings = {
    oauth = {
      # The web type grants authorization_code and password by default: keep the authorization code flow only.
      grant_types               = ["authorization_code", "refresh_token"]
      redirect_uris             = ["https://shop.example.com/login/callback"]
      post_logout_redirect_uris = ["https://shop.example.com/"]
      scope_settings = [
        { scope = "openid", default_scope = true },
        { scope = "profile", default_scope = true },
        { scope = "email" },
      ]
    }
  }
}

# The website authenticates with its client ID and this secret (client_secret_basic).
resource "am_client_secret" "web" {
  domain_key      = am_domain.shop.key
  application_key = am_application.web.key
  key             = "web-2026-11"
  value_wo        = var.client_secret
}

output "client_id" {
  value = am_application.web.client_id
}

variable "keystore_password" {
  type      = string
  sensitive = true
}

variable "key_password" {
  type      = string
  sensitive = true
}

variable "alice_password" {
  type      = string
  sensitive = true
}

# Read from your secret store and given to the website too, for example with TF_VAR_client_secret.
variable "client_secret" {
  type      = string
  sensitive = true
  ephemeral = true
}

```

## How it works

* The domain overrides the organization's login settings (`inherited = false`): its login page offers
  registration and password reset. Settings blocks left out of the configuration keep AM's values.
* The `web` type is a confidential client: it authenticates at the token endpoint with its client ID and secret
  (`client_secret_basic`). Its default grant types are `authorization_code` and `password`; declaring
  `grant_types` keeps the authorization code flow and refresh tokens only.
* `client_id` is generated by AM when the configuration leaves it out, and exported as an output; give it to
  the website with the secret. Set `client_id` to choose it instead, as the mobile application of the PKCE guide does.
* AM never returns the secret: `value_wo` is sent on creation only, and never stored in the plan or the state.
  Feed it from your secret store, so the website reads the same value.
* The OpenID Connect discovery document of the domain is served by the AM gateway at
  `<gateway URL>/acme-shop/oidc/.well-known/openid-configuration`.
