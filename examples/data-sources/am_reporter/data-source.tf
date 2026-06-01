data "am_reporter" "my_reporter" {
  domain_key      = "customers"
  environment_id  = "DEFAULT"
  key             = "audit-kafka"
  organization_id = "DEFAULT"
}