---
page_title: "Service application and client secret rotation"
subcategory: "Applications"
---

# Service application and client secret rotation

This guide creates a backend service that gets access tokens on its own behalf (OAuth 2.0 client credentials),
with a client secret that expires, and shows how to rotate that secret without downtime.

## Before you start

* The provider reaches the AM Automation API with `AM_SERVER_URL` and a service-account token in `AM_SA_TOKEN`.
* Terraform 1.11 or later: the client secret value is a write-only attribute, never stored in the state.
* A Java keystore holding the RSA key that signs the tokens, next to the configuration:

```shell
keytool -genkeypair -alias signing -keyalg RSA -keysize 2048 -validity 365 -dname CN=acme-apis \
  -storetype JKS -keystore keystore.jks -storepass "$TF_VAR_keystore_password" -keypass "$TF_VAR_key_password"
```

## Configuration

```terraform
# A backend service calling APIs on its own behalf (client credentials flow), with an expiring client secret.

resource "am_domain" "apis" {
  key           = "acme-apis"
  name          = "ACME APIs"
  description   = "Machine-to-machine access to the ACME APIs"
  path          = "/acme-apis"
  data_plane_id = "default"
}

# Signs the access tokens. keystore.jks holds an RSA key pair under the alias "signing".
resource "am_certificate" "signing" {
  domain_key = am_domain.apis.key
  key        = "apis-signing"
  name       = "APIs signing key"
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

# The service type grants client_credentials only.
resource "am_application" "billing" {
  domain_key  = am_domain.apis.key
  key         = "billing-service"
  name        = "Billing service"
  type        = "service"
  certificate = am_certificate.signing.key

  settings = {
    oauth = {
      access_token_validity_seconds = 900
    }
    # Secrets of this application expire 90 days after their creation: rotate them before.
    secret_expiration_settings = {
      enabled             = true
      expiry_time_seconds = 7776000
    }
  }
}

# A secret is immutable. To rotate it, add a second am_client_secret with a new key, apply, move the
# service to the new value, then remove this one and apply again.
resource "am_client_secret" "billing_2026_11" {
  domain_key      = am_domain.apis.key
  application_key = am_application.billing.key
  key             = "billing-2026-11"
  value_wo        = var.client_secret
}

output "client_id" {
  value = am_application.billing.client_id
}

output "client_secret_expires_at" {
  value = am_client_secret.billing_2026_11.expires_at
}

variable "keystore_password" {
  type      = string
  sensitive = true
}

variable "key_password" {
  type      = string
  sensitive = true
}

# Read from your secret store and given to the service too, for example with TF_VAR_client_secret.
variable "client_secret" {
  type      = string
  sensitive = true
  ephemeral = true
}

```

Once applied, the service gets a token from the AM gateway:

```shell
curl -u "$(terraform output -raw client_id):$TF_VAR_client_secret" \
  -d grant_type=client_credentials <gateway URL>/acme-apis/oauth/token
```

## Rotate the secret

A client secret is immutable and `value_wo` is only sent on creation: changing the value of an existing
`am_client_secret` has no effect. Rotate it with two applies instead.

1. Add a secret with a new `key` and its own value, then apply. Both secrets are valid.

   ```terraform
   resource "am_client_secret" "billing_2027_02" {
     domain_key      = am_domain.apis.key
     application_key = am_application.billing.key
     key             = "billing-2027-02"
     value_wo        = var.next_client_secret
   }
   ```

2. Move the service to the new value.
3. Remove `am_client_secret.billing_2026_11` from the configuration and apply: AM revokes the old secret.

Changing the `key` of an existing secret is not a rotation: Terraform replaces the secret in a single apply and
the old value stops working at once.

The `expires_at` attribute and output tell when a secret expires, from the `secret_expiration_settings` of the
application, else of the domain.
