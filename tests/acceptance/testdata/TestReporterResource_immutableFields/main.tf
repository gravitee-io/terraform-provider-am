variable "domain_key" {
  type = string
}

variable "data_plane_id" {
  type = string
}

variable "key" {
  type = string
}

variable "type" {
  type = string
}

variable "system" {
  type    = bool
  default = null
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
  type       = var.type
  system     = var.system
  enabled    = true
  configuration = jsonencode({
    filename = var.key
  })
}
