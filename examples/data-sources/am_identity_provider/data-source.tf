data "am_identity_provider" "my_identityprovider" {
  domain_key = "customers"
  env_id     = "DEFAULT"
  key        = "corporate-ldap"
  org_id     = "DEFAULT"
}