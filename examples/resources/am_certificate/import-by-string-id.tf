import {
  to = am_certificate.my_am_certificate
  id = jsonencode({
    domain_key      = "customers"
    environment_id  = "DEFAULT"
    key             = "signing-cert"
    organization_id = "DEFAULT"
  })
}
