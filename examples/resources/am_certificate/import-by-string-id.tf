import {
  to = am_certificate.signing
  id = jsonencode({
    domain_key      = "example-domain"
    environment_id  = "DEFAULT"
    key             = "signing-cert"
    organization_id = "DEFAULT"
  })
}
