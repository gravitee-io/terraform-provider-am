variable "domain_key" {
  type = string
}

variable "data_plane_id" {
  type = string
}

variable "name" {
  type = string
}

resource "am_domain" "test" {
  key           = var.domain_key
  name          = var.name
  path          = "/${var.domain_key}"
  data_plane_id = var.data_plane_id
}
