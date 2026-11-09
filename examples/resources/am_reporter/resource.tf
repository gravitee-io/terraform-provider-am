# Writes the domain's audit events to a file on the AM nodes.
resource "am_reporter" "audit" {
  domain_key = "example-domain"
  key        = "audit-file"
  name       = "Audit events to a file"
  type       = "reporter-am-file"
  configuration = jsonencode({
    filename = "example-domain-audit"
  })
}
