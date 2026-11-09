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

# The certificate does not exist: AM accepts a reference to one created later.
resource "am_application" "test" {
  key         = var.key
  domain_key  = am_domain.test.key
  name        = "Terraform acceptance"
  type        = "native"
  certificate = "created-later"

  settings = {
    oauth = {
      redirect_uris = ["https://app.example.com/callback"]
    }
  }
}

output "client_id" {
  value = am_application.test.client_id
}
