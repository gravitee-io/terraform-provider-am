data "am_certificate" "my_certificate" {
  domain_key      = "customers"
  environment_id  = "...my_environment_id..."
  key             = "signing-cert"
  organization_id = "...my_organization_id..."
}