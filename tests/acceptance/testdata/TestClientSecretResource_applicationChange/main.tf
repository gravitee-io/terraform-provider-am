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

variable "second_application_key" {
  type = string
}

variable "in_second_application" {
  type = bool
}

resource "am_domain" "test" {
  key           = var.domain_key
  name          = "Terraform acceptance"
  path          = "/${var.domain_key}"
  data_plane_id = var.data_plane_id
}

resource "am_application" "first" {
  key         = var.application_key
  domain_key  = am_domain.test.key
  name        = "Terraform acceptance"
  type        = "service"
  certificate = "created-later"
}

resource "am_application" "second" {
  key         = var.second_application_key
  domain_key  = am_domain.test.key
  name        = "Terraform acceptance"
  type        = "service"
  certificate = "created-later"
}

ephemeral "random_password" "test" {
  length  = 64
  special = false
}

resource "am_client_secret" "test" {
  domain_key      = am_domain.test.key
  application_key = var.in_second_application ? am_application.second.key : am_application.first.key
  key             = "secret-2026-10"
  value_wo        = ephemeral.random_password.test.result
}
