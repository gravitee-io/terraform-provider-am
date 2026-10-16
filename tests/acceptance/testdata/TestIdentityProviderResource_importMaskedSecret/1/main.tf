variable "domain_key" {
  type = string
}

variable "data_plane_id" {
  type = string
}

resource "am_domain" "test" {
  key           = var.domain_key
  name          = "Terraform acceptance"
  path          = "/${var.domain_key}"
  data_plane_id = var.data_plane_id
}
