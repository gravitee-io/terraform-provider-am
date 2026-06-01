resource "am_identity_provider" "my_identityprovider" {
  configuration = "{\"users\":[{\"username\":\"admin\",\"password\":\"...\"}]}"
  domain_key    = "customers"
  domain_whitelist = [
    "example.com",
  ]
  env_id = "DEFAULT"
  group_mapper = {
    key = [
      # ...
    ]
  }
  key = "corporate-ldap"
  mappers = {
    key = "value"
  }
  name   = "Corporate LDAP"
  org_id = "DEFAULT"
  role_mapper = {
    key = [
      # ...
    ]
  }
  system = false
  type   = "inline-am-idp"
}