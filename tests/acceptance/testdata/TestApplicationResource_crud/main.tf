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

variable "name" {
  type = string
}

variable "redirect_uri" {
  type = string
}

variable "owner" {
  type = string
}

variable "keystore" {
  type = string
}

resource "am_certificate" "test" {
  domain_key = am_domain.test.key
  key        = "signing"
  name       = "Terraform acceptance"
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

resource "am_identity_provider" "test" {
  domain_key = am_domain.test.key
  key        = "users"
  name       = "Terraform acceptance users"
  type       = "inline-am-idp"
  configuration = jsonencode({
    users = [{
      firstname = "Alice"
      lastname  = "Acceptance"
      username  = "alice"
      password  = "Tf-Acc-P@ssw0rd"
    }]
  })
}

resource "am_application" "test" {
  key         = var.key
  domain_key  = am_domain.test.key
  name        = var.name
  description = "Customer mobile client"
  type        = "native"
  certificate = am_certificate.test.key

  metadata = {
    owner = jsonencode(var.owner)
  }

  settings = {
    oauth = {
      grant_types   = ["authorization_code", "refresh_token"]
      redirect_uris = [var.redirect_uri]
      scope_settings = [
        { scope = "openid", default_scope = true },
      ]
    }
  }

  identity_providers = [
    { identity = am_identity_provider.test.key, priority = 0 },
  ]
}
