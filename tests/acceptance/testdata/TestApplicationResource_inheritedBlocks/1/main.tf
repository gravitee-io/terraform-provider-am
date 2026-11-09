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

# No login block: inherited from the domain.
resource "am_application" "test" {
  key         = var.key
  domain_key  = am_domain.test.key
  name        = "Terraform acceptance"
  type        = "web"
  certificate = "created-later"

  settings = {
    oauth = {
      redirect_uris = ["https://app.example.com/callback"]
    }
  }
}
