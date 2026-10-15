variable "domain_key" {
  type = string
}

variable "data_plane_id" {
  type = string
}

variable "key" {
  type = string
}

variable "enabled" {
  type = bool
}

resource "am_domain" "test" {
  key           = var.domain_key
  name          = "Terraform acceptance"
  path          = "/${var.domain_key}"
  data_plane_id = var.data_plane_id
}

resource "am_reporter" "test" {
  domain_key = am_domain.test.key
  key        = var.key
  name       = "Terraform acceptance audit"
  type       = "reporter-am-file"
  enabled    = var.enabled
  configuration = jsonencode({
    filename = var.key
  })
}
