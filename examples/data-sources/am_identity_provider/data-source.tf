data "am_identity_provider" "my_identityprovider" {
  domain_key      = "customers"
  environment_id  = "DEFAULT"
  key             = "corporate-ldap"
  organization_id = "DEFAULT"
}