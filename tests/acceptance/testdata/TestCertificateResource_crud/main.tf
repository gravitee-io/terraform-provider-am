variable "domain_key" {
  type = string
}

variable "data_plane_id" {
  type = string
}

variable "key" {
  type = string
}

variable "name" {
  type = string
}

variable "keystore" {
  type = string
}

resource "am_domain" "test" {
  key           = var.domain_key
  name          = "Terraform acceptance"
  path          = "/${var.domain_key}"
  data_plane_id = var.data_plane_id
}

resource "am_certificate" "test" {
  domain_key = am_domain.test.key
  key        = var.key
  name       = var.name
  type       = "javakeystore-am-certificate"
  configuration = jsonencode({
    jks = jsonencode({
      name    = "server.jks"
      type    = ""
      size    = 2237
      content = var.keystore
    })
    storepass = "letmein"
    alias     = "mytestkey"
    keypass   = "changeme"
  })
}
