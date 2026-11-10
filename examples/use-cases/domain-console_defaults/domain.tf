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
