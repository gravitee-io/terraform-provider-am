data "am_reporter" "my_reporter" {
  domain_key      = "customers"
  environment_id  = "...my_environment_id..."
  key             = "audit-kafka"
  organization_id = "...my_organization_id..."
}