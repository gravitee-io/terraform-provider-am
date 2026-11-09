variable "domain_key" {
  type = string
}

variable "data_plane_id" {
  type = string
}

variable "application_key" {
  type = string
}

resource "am_domain" "test" {
  key           = var.domain_key
  name          = "Terraform acceptance"
  path          = "/${var.domain_key}"
  data_plane_id = var.data_plane_id
}

resource "am_application" "test" {
  key         = var.application_key
  domain_key  = am_domain.test.key
  name        = "Terraform acceptance"
  type        = "service"
  certificate = "created-later"

  # Gives the secret an expiry date.
  settings = {
    secret_expiration_settings = {
      enabled             = true
      expiry_time_seconds = 86400
    }
  }
}

# Ephemeral, so the test knows the value and can check it never reaches the state.
variable "value" {
  type      = string
  ephemeral = true
}

resource "am_client_secret" "test" {
  domain_key      = am_domain.test.key
  application_key = am_application.test.key
  key             = "secret-2026-10"
  value_wo        = var.value
}
