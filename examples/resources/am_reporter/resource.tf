resource "am_reporter" "my_reporter" {
  configuration = "{\"bootstrapServers\":\"kafka:9092\",\"topic\":\"audit\"}"
  domain_key    = "customers"
  enabled       = true
  env_id        = "DEFAULT"
  key           = "audit-kafka"
  name          = "Audit events to Kafka"
  org_id        = "DEFAULT"
  system        = false
  type          = "reporter-am-kafka"
}