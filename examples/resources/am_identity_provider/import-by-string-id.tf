import {
  to = am_identity_provider.my_am_identity_provider
  id = jsonencode({
    domain_key      = "customers"
    environment_id  = "..."
    key             = "corporate-ldap"
    organization_id = "..."
  })
}
