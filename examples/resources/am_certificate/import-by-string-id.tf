import {
  to = am_certificate.my_am_certificate
  id = jsonencode({
    domain_key      = "customers"
    environment_id  = "..."
    key             = "signing-cert"
    organization_id = "..."
  })
}
