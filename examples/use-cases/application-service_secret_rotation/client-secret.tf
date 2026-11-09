# A secret is immutable. To rotate it, add a second am_client_secret with a new key, apply, move the
# service to the new value, then remove this file and apply again.
resource "am_client_secret" "billing_2026_11" {
  domain_key      = am_domain.apis.key
  application_key = am_application.billing.key
  key             = "billing-2026-11"
  value_wo        = var.client_secret
}

output "billing_2026_11_expires_at" {
  value = am_client_secret.billing_2026_11.expires_at
}

# Read from your secret store and given to the service too, for example with TF_VAR_client_secret.
variable "client_secret" {
  type      = string
  sensitive = true
  ephemeral = true
}
