resource "am_reporter" "my_reporter" {
  attribute_mapping_event_types = [
    "..."
  ]
  attribute_mappings = [
    {
      exported_name = "user_sub"
      expression    = "{#context.attributes['user'].additionalInformation['sub']}"
    }
  ]
  configuration   = "{\"bootstrapServers\":\"kafka:9092\",\"topic\":\"audit\"}"
  domain_key      = "example-domain"
  enabled         = true
  environment_id  = "DEFAULT"
  key             = "audit-kafka"
  name            = "Audit events to Kafka"
  organization_id = "DEFAULT"
  system          = false
  type            = "reporter-am-kafka"
}