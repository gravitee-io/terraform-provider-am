variable "domain_key" {
  type = string
}

variable "data_plane_id" {
  type = string
}

variable "key" {
  type = string
}

resource "am_domain" "test" {
  key           = var.domain_key
  name          = "Terraform acceptance"
  path          = "/${var.domain_key}"
  data_plane_id = var.data_plane_id
}

variable "type" {
  type = string
}

variable "token_endpoint_auth_method" {
  type    = string
  default = null
}

variable "grant_types" {
  type    = list(string)
  default = null
}

resource "am_application" "test" {
  key         = var.key
  domain_key  = am_domain.test.key
  name        = "Terraform acceptance"
  type        = var.type
  certificate = "created-later"

  settings = {
    oauth = {
      redirect_uris              = ["https://app.example.com/callback"]
      token_endpoint_auth_method = var.token_endpoint_auth_method
      grant_types                = var.grant_types
    }
  }
}
