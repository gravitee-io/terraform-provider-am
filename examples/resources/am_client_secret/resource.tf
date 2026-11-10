resource "am_client_secret" "mobile_app_2026_10" {
  domain_key      = "example-domain"
  application_key = "mobile-app"
  key             = "secret-2026-10"
  value_wo        = var.client_secret
}

# AM never returns the value: read it from the secret store the clients read it from, for example with
# TF_VAR_client_secret.
variable "client_secret" {
  type      = string
  sensitive = true
  ephemeral = true
}
