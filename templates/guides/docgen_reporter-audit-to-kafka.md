---
page_title: "Send audit events to Kafka"
subcategory: "Reporters"
---

# Send audit events to Kafka

This guide sends the audit events of a domain to Kafka, for example for a SIEM, and keeps the audit log of the
console.

## Before you start

* The provider reaches the AM Automation API with `AM_SERVER_URL` and a service-account token in `AM_SA_TOKEN`.
* A Kafka cluster AM reaches, and the password of its user in `TF_VAR_kafka_password`.

## Configuration

```terraform
# The audit events of a domain sent to Kafka, for the security team's SIEM, on top of the console's audit log.

resource "am_domain" "payments" {
  key           = "acme-payments"
  name          = "ACME payments"
  path          = "/acme-payments"
  data_plane_id = "default"
}

# The console's audit log reads the system reporter, which a domain created by Terraform does not get.
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

```

## How it differs from the console

* The audit log of the console reads the domain's system reporter, which a domain created by Terraform does not
  get: `am_reporter.console` declares it, with `system = true` and its key only.
* AM returns the secrets of a plugin configuration as `********`. The provider compares the configuration
  with that masking taken into account and keeps the configured values in the state: a secret changed in the
  console does not show in the plan. Change the variable instead, and apply.
* After an import, the state holds only `********`: the first plan shows the configuration as changed, and the
  apply sends the configured secrets again.
* `attribute_mappings` adds fields to the exported events: each expression reads the audit context
  (`user`, `client`, `request`, `audit`), and its value is exported under `exported_name`.
  `attribute_mapping_event_types` limits them to these event types; empty means every event type. A system
  reporter ignores both.
* `type` and `system` cannot change once created: AM rejects the apply.
