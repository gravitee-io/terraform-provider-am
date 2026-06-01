data "am_certificate" "my_certificate" {
  domain_key      = "customers"
  environment_id  = "DEFAULT"
  key             = "signing-cert"
  organization_id = "DEFAULT"
}