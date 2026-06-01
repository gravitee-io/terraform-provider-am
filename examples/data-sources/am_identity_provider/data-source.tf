data "am_identity_provider" "my_identityprovider" {
  domain_key      = "customers"
  environment_id  = "...my_environment_id..."
  key             = "corporate-ldap"
  organization_id = "...my_organization_id..."
}