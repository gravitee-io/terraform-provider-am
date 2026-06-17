data "am_reporter" "my_reporter" {
  domain_key      = "example-domain"
  environment_id  = "DEFAULT"
  key             = "audit-kafka"
  organization_id = "DEFAULT"
}