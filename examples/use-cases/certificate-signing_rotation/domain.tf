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
    }
  }
}

variable "keystore_password" {
  type      = string
  sensitive = true
}

variable "key_password" {
  type      = string
  sensitive = true
}
