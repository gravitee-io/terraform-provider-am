resource "am_identity_provider" "my_identityprovider" {
  configuration = "{\"users\":[{\"username\":\"admin\",\"password\":\"...\"}]}"
  domain_key    = "example-domain"
  domain_whitelist = [
    "example.com",
  ]
  environment_id = "DEFAULT"
  external       = false
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
  organization_id = "DEFAULT"
  role_mapper = {
    key = [
      # ...
    ]
  }
  system = false
  type   = "inline-am-idp"
}