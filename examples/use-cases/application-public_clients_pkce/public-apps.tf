# A single-page application and a mobile application: public clients, which cannot keep a secret, so
# they sign users in with the authorization code flow and PKCE.

resource "am_domain" "app" {
  key           = "acme-app"
  name          = "ACME app"
  description   = "Users of the ACME web and mobile apps"
  path          = "/acme-app"
  data_plane_id = "default"
}

# Signs the ID and access tokens. keystore.jks holds an RSA key pair under the alias "signing".
resource "am_certificate" "signing" {
  domain_key = am_domain.app.key
  key        = "app-signing"
  name       = "App signing key"
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

resource "am_identity_provider" "users" {
  domain_key = am_domain.app.key
  key        = "app-users"
  name       = "App users"
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

# browser and native applications created without grant_types get the authorization code flow, PKCE
# with S256 and no client authentication (token_endpoint_auth_method = "none").
resource "am_application" "spa" {
  domain_key  = am_domain.app.key
  key         = "app-spa"
  name        = "ACME single-page app"
  type        = "browser"
  certificate = am_certificate.signing.key

  identity_providers = [
    { identity = am_identity_provider.users.key }
  ]

  settings = {
    oauth = {
      redirect_uris             = ["https://app.example.com/callback"]
      post_logout_redirect_uris = ["https://app.example.com/"]
      scope_settings = [
        { scope = "openid", default_scope = true },
        { scope = "profile", default_scope = true },
      ]
    }
  }
}

resource "am_application" "mobile" {
  domain_key  = am_domain.app.key
  key         = "app-mobile"
  name        = "ACME mobile app"
  type        = "native"
  certificate = am_certificate.signing.key

  identity_providers = [
    { identity = am_identity_provider.users.key }
  ]

  settings = {
    oauth = {
      redirect_uris = ["com.example.app:/oauth2/callback"]
      scope_settings = [
        { scope = "openid", default_scope = true },
        { scope = "profile", default_scope = true },
      ]
    }
  }
}

output "spa_client_id" {
  value = am_application.spa.client_id
}

output "mobile_client_id" {
  value = am_application.mobile.client_id
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
