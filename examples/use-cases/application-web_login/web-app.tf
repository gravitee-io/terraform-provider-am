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
