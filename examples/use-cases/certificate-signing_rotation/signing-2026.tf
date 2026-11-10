# The current signing certificate, from keystore-2026.jks.
resource "am_certificate" "signing_2026" {
  domain_key = am_domain.accounts.key
  key        = "signing-2026"
  name       = "Signing key 2026"
  type       = "javakeystore-am-certificate"
  configuration = jsonencode({
    jks = jsonencode({
      name    = "keystore-2026.jks"
      content = filebase64("${path.module}/keystore-2026.jks")
    })
    storepass = var.keystore_password
    alias     = "signing"
    keypass   = var.key_password
  })
}

output "signing_2026_expires_at" {
  value = am_certificate.signing_2026.expires_at
}
