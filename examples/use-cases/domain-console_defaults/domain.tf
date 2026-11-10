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
