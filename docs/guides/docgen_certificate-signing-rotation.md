---
page_title: "Token signing certificate rotation"
subcategory: "Certificates"
---

# Token signing certificate rotation

This guide rotates the certificate that signs the tokens of a domain before it expires, without breaking the
tokens it already signed.

## Before you start

* The provider reaches the AM Automation API with `AM_SERVER_URL` and a service-account token in `AM_SA_TOKEN`.
* A Java keystore holding the RSA key that signs the tokens, next to the configuration, its passwords in
  `TF_VAR_keystore_2026_password` and `TF_VAR_key_2026_password`:

```shell
keytool -genkeypair -alias signing -keyalg RSA -keysize 2048 -validity 365 -dname CN=acme-accounts \
  -storetype JKS -keystore keystore-2026.jks \
  -storepass:env TF_VAR_keystore_2026_password -keypass:env TF_VAR_key_2026_password
```

## Configuration

The domain and its application, in `domain.tf`:

```terraform
# A domain whose token-signing certificate is rotated before it expires, without invalidating the tokens
# already issued.

resource "am_domain" "accounts" {
  key           = "acme-accounts"
  name          = "ACME accounts"
  path          = "/acme-accounts"
  data_plane_id = "default"

  # Signs the tokens of the applications that name no certificate.
  certificate_settings = {
    fallback_certificate = local.signing_certificate
  }
}

resource "am_application" "account_portal" {
  domain_key  = am_domain.accounts.key
  key         = "account-portal"
  name        = "Account portal"
  type        = "web"
  certificate = local.signing_certificate

  settings = {
    oauth = {
      redirect_uris = ["https://accounts.example.com/callback"]
      # The longest of these validities is how long to keep the old certificate after the switch.
      access_token_validity_seconds  = 3600
      id_token_validity_seconds      = 3600
      refresh_token_validity_seconds = 86400
    }
  }
}

```

The current certificate, in `signing-2026.tf`:

```terraform
# The current signing certificate, from keystore-2026.jks, with its own passwords.
resource "am_certificate" "signing_2026" {
  domain_key = am_domain.accounts.key
  key        = "signing-2026"
  name       = "Signing key 2026"
  type       = "javakeystore-am-certificate"
  configuration = jsonencode({
    jks = jsonencode({
      name    = "keystore-2026.jks"
      content = filebase64("${path.module}/keystore-2026.jks")
    })
    storepass = var.keystore_2026_password
    alias     = "signing"
    keypass   = var.key_2026_password
  })
}

output "signing_2026_expires_at" {
  value = am_certificate.signing_2026.expires_at
}

variable "keystore_2026_password" {
  type      = string
  sensitive = true
}

variable "key_2026_password" {
  type      = string
  sensitive = true
}

```

The certificate in use, in `current-certificate.tf`:

```terraform
# The certificate signing the tokens, named by its key: the domain and the applications follow it.
locals {
  signing_certificate = "signing-2026"
}

```

## Rotate the certificate

1. Create the next keystore with its own passwords, in `TF_VAR_keystore_2027_password` and
   `TF_VAR_key_2027_password`:

```shell
keytool -genkeypair -alias signing -keyalg RSA -keysize 2048 -validity 365 -dname CN=acme-accounts \
  -storetype JKS -keystore keystore-2027.jks \
  -storepass:env TF_VAR_keystore_2027_password -keypass:env TF_VAR_key_2027_password
```

   Add the next certificate in a new file, for example `signing-2027.tf`, and apply. Nothing signs with it yet.

```terraform
# The next signing certificate, from keystore-2027.jks, with its own passwords.
resource "am_certificate" "signing_2027" {
  domain_key = am_domain.accounts.key
  key        = "signing-2027"
  name       = "Signing key 2027"
  type       = "javakeystore-am-certificate"
  configuration = jsonencode({
    jks = jsonencode({
      name    = "keystore-2027.jks"
      content = filebase64("${path.module}/keystore-2027.jks")
    })
    storepass = var.keystore_2027_password
    alias     = "signing"
    keypass   = var.key_2027_password
  })
}

output "signing_2027_expires_at" {
  value = am_certificate.signing_2027.expires_at
}

variable "keystore_2027_password" {
  type      = string
  sensitive = true
}

variable "key_2027_password" {
  type      = string
  sensitive = true
}

```

2. Name it in `current-certificate.tf`, and apply: the domain and the application sign with it.

```terraform
# The certificate signing the tokens, named by its key: the domain and the applications follow it.
locals {
  signing_certificate = "signing-2027"
}

```

3. Wait for the tokens signed by the old key to expire: the longest token validity of the applications that used
   it, here `refresh_token_validity_seconds`, one day. Then delete `signing-2026.tf` and `keystore-2026.jks`, unset
   the 2026 passwords, and apply.

Until step 3, the old certificate stays in the domain: AM still accepts the access and refresh tokens it signed,
and the domain's JWKS (`<gateway URL>/acme-accounts/oidc/.well-known/jwks.json`) publishes both public keys, so the
resource servers verify the tokens signed by either.

## How it differs from the console

* The domain and the applications name a certificate by its key, as a plain string here so that one local value
  switches them all. AM does not check the reference: it resolves whenever a certificate with that key exists.
* AM does not stop you from deleting a certificate that the domain or an application still names: their
  reference stays, and resolves again once a certificate with that key exists. Switch them first, as in step 2.
* Changing the keystore of an existing certificate is not a rotation: the new key signs at once, and the
  tokens signed by the old one fail to verify.
* AM returns the keystore and its passwords as `********`; the provider keeps the configured values.
* `expires_at` and the outputs tell when a certificate expires.
