# The next signing certificate, from a new keystore.
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
    storepass = var.keystore_password
    alias     = "signing"
    keypass   = var.key_password
  })
}

output "signing_2027_expires_at" {
  value = am_certificate.signing_2027.expires_at
}
