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
