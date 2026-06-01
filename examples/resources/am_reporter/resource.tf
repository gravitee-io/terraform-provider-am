resource "am_reporter" "my_reporter" {
  configuration   = "{\"bootstrapServers\":\"kafka:9092\",\"topic\":\"audit\"}"
  domain_key      = "customers"
  enabled         = true
  environment_id  = "...my_environment_id..."
  key             = "audit-kafka"
  name            = "Audit events to Kafka"
  organization_id = "...my_organization_id..."
  system          = false
  type            = "reporter-am-kafka"
}