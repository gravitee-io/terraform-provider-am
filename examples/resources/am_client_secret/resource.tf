ephemeral "random_password" "mobile_app_2026_10" {
  length  = 64
  special = false
}

resource "am_client_secret" "mobile_app_2026_10" {
  domain_key      = "example-domain"
  application_key = "mobile-app"
  key             = "secret-2026-10"
  value_wo        = ephemeral.random_password.mobile_app_2026_10.result
}
