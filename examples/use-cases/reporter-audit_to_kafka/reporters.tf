# The audit events of a domain sent to Kafka, for the security team's SIEM, on top of the console's audit log
# and dashboard.

resource "am_domain" "payments" {
  key           = "acme-payments"
  name          = "ACME payments"
  path          = "/acme-payments"
  data_plane_id = "default"
}

# The console's audit log and dashboard read a reporter storing the events in a database, not the Kafka one.
# A domain created by Terraform gets none: this is the system one, built from AM's defaults.
resource "am_reporter" "console" {
  domain_key = am_domain.payments.key
  key        = "default"
  system     = true
}

resource "am_reporter" "siem" {
  domain_key = am_domain.payments.key
  key        = "siem-kafka"
  name       = "SIEM (Kafka)"
  type       = "reporter-am-kafka"

  # AM returns password as ********: the provider keeps the configured value instead of seeing a change.
  configuration = jsonencode({
    bootstrapServers = "kafka-1.example.com:9092,kafka-2.example.com:9092"
    topic            = "am-audit-payments"
    acks             = "all"
    username         = "am-audit"
    password         = var.kafka_password
  })

  # Adds the user ID and the client IP to the sign-in and sign-out events.
  attribute_mapping_event_types = ["USER_LOGIN", "USER_LOGOUT"]
  attribute_mappings = [
    { expression = "{#context.attributes['user'].id}", exported_name = "user_id" },
    { expression = "{#context.attributes['request'].ip}", exported_name = "client_ip" },
  ]
}

variable "kafka_password" {
  type      = string
  sensitive = true
}
