# A web application signing its tokens with the domain's certificate and authenticating users
# with its identity provider. client_id is generated when omitted; am_client_secret gives it a secret.
resource "am_application" "mobile_app" {
  domain_key  = "example-domain"
  key         = "mobile-app"
  name        = "Mobile app"
  description = "Customer mobile application"
  type        = "web"
  certificate = "signing-cert"

  identity_providers = [
    {
      identity = "customer-users"
      priority = 0
    }
  ]

  settings = {
    oauth = {
      redirect_uris = ["https://app.example.com/callback"]
      scope_settings = [
        { scope = "openid", default_scope = true },
        { scope = "profile" },
      ]
      access_token_validity_seconds = 3600
    }
    login = {
      inherited               = false
      forgot_password_enabled = true
    }
    secret_expiration_settings = {
      enabled             = true
      expiry_time_seconds = 2592000 # 30 days
    }
  }

  metadata = {
    owner = jsonencode("payments-team")
  }
}

output "mobile_app_client_id" {
  value = am_application.mobile_app.client_id
}
