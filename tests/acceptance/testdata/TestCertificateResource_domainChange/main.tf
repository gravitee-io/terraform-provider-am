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

variable "name" {
  type = string
}

variable "keystore" {
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

resource "am_certificate" "test" {
  domain_key = var.in_second_domain ? am_domain.second.key : am_domain.first.key
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
