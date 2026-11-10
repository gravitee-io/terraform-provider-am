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
