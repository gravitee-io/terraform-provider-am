# The next signing certificate, from keystore-2027.jks, with its own passwords.
resource "am_certificate" "signing_2027" {
  domain_key = am_domain.accounts.key
  key        = "signing-2027"
  name       = "Signing key 2027"
  type       = "javakeystore-am-certificate"
  configuration = jsonencode({
    jks = jsonencode({
      name    = "keystore-2027.jks"
      content = filebase64("${path.module}/keystore-2027.jks")
    })
    storepass = var.keystore_2027_password
    alias     = "signing-2027"
    keypass   = var.key_2027_password
  })
}

output "signing_2027_expires_at" {
  value = am_certificate.signing_2027.expires_at
}

variable "keystore_2027_password" {
  type      = string
  sensitive = true
}

variable "key_2027_password" {
  type      = string
  sensitive = true
}
