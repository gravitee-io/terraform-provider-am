resource "am_identity_provider" "my_identityprovider" {
  configuration = "{\"users\":[{\"username\":\"admin\",\"password\":\"...\"}]}"
  domain_key    = "customers"
  domain_whitelist = [
    "example.com",
  ]
  environment_id = "...my_environment_id..."
  group_mapper = {
    key = [
      # ...
    ]
  }
  key = "corporate-ldap"
  mappers = {
    key = "value"
  }
  name            = "Corporate LDAP"
  organization_id = "...my_organization_id..."
  role_mapper = {
    key = [
      # ...
    ]
  }
  system = false
  type   = "inline-am-idp"
}