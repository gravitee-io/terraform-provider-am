import {
  to = am_reporter.my_am_reporter
  id = jsonencode({
    domain_key      = "customers"
    environment_id  = "DEFAULT"
    key             = "audit-kafka"
    organization_id = "DEFAULT"
  })
}
