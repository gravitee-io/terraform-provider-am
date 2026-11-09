variable "domain_key" {
  type = string
}

variable "second_domain_key" {
  type = string
}

variable "in_second_domain" {
  type = bool
}

variable "data_plane_id" {
  type = string
}

variable "key" {
  type = string
}

resource "am_domain" "first" {
  key           = var.domain_key
  name          = var.domain_key
  path          = "/${var.domain_key}"
  data_plane_id = var.data_plane_id
}

resource "am_domain" "second" {
  key           = var.second_domain_key
  name          = var.second_domain_key
  path          = "/${var.second_domain_key}"
  data_plane_id = var.data_plane_id
}

resource "am_application" "test" {
  domain_key  = var.in_second_domain ? am_domain.second.key : am_domain.first.key
  key         = var.key
  name        = "Terraform acceptance"
  type        = "service"
  certificate = "created-later"
}
