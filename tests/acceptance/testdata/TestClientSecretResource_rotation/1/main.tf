terraform {
  required_providers {
    random = {
      source  = "hashicorp/random"
      version = ">= 3.7.0"
    }
  }
}

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
}


ephemeral "random_password" "secret_2026_10" {
  length  = 64
  special = false
}

resource "am_client_secret" "secret_2026_10" {
  domain_key      = am_domain.test.key
  application_key = am_application.test.key
  key             = "secret-2026-10"
  value_wo        = ephemeral.random_password.secret_2026_10.result
}
