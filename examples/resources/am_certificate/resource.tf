# A Java keystore holding the key that signs the domain's tokens. AM masks the keystore and its
# passwords in its responses; the provider keeps the configured values.
resource "am_certificate" "signing" {
  domain_key = "example-domain"
  key        = "signing-cert"
  name       = "Signing certificate"
  type       = "javakeystore-am-certificate"
  configuration = jsonencode({
    jks = jsonencode({
      name    = "keystore.jks"
      content = filebase64("${path.module}/keystore.jks")
    })
    storepass = var.keystore_password
    alias     = "signing"
    keypass   = var.key_password
  })
}

variable "keystore_password" {
  type      = string
  sensitive = true
}

variable "key_password" {
  type      = string
  sensitive = true
}
