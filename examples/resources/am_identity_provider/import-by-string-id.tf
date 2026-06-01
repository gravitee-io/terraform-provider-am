import {
  to = am_identity_provider.my_am_identity_provider
  id = jsonencode({
    domain_key = "customers"
    env_id     = "DEFAULT"
    key        = "corporate-ldap"
    org_id     = "DEFAULT"
  })
}
