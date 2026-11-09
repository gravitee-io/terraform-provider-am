# An inline identity provider: users declared in the configuration, handy for tests and demos.
resource "am_identity_provider" "customers" {
  domain_key = "example-domain"
  key        = "customer-users"
  name       = "Customer users"
  type       = "inline-am-idp"
  configuration = jsonencode({
    users = [{
      firstname = "Alice"
      lastname  = "Example"
      username  = "alice"
      password  = var.alice_password
    }]
  })
}

variable "alice_password" {
  type      = string
  sensitive = true
}
