import {
  to = am_identity_provider.customers
  id = jsonencode({
    domain_key      = "example-domain"
    environment_id  = "DEFAULT"
    key             = "customer-users"
    organization_id = "DEFAULT"
  })
}
