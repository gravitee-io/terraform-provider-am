# The secret that replaces billing_2026_11: applied next to it, so both are valid until the service moves.
resource "am_client_secret" "billing_2027_02" {
  domain_key      = am_domain.apis.key
  application_key = am_application.billing.key
  key             = "billing-2027-02"
  value_wo        = var.next_client_secret
}

output "billing_2027_02_expires_at" {
  value = am_client_secret.billing_2027_02.expires_at
}

variable "next_client_secret" {
  type      = string
  sensitive = true
  ephemeral = true
}
