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

output "client_id" {
  value = am_application.billing.client_id
}

variable "keystore_password" {
  type      = string
  sensitive = true
}

variable "key_password" {
  type      = string
  sensitive = true
}
