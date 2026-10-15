variable "domain_key" {
  type = string
}

variable "data_plane_id" {
  type = string
}

variable "organization_id" {
  type = string
}

variable "environment_id" {
  type = string
}

variable "key" {
  type = string
}

variable "password" {
  type = string
}

resource "am_domain" "test" {
  key           = var.domain_key
  name          = "Terraform acceptance"
  path          = "/${var.domain_key}"
  data_plane_id = var.data_plane_id
}

resource "am_identity_provider" "test" {
  domain_key = am_domain.test.key
  key        = var.key
  name       = "Terraform acceptance users"
  type       = "inline-am-idp"
  configuration = jsonencode({
    users = [{
      firstname = "Alice"
      lastname  = "Acceptance"
      username  = "alice"
      password  = var.password
    }]
  })
}

import {
  to = am_identity_provider.test
  id = jsonencode({ domain_key = var.domain_key, environment_id = var.environment_id, key = var.key, organization_id = var.organization_id })
}
