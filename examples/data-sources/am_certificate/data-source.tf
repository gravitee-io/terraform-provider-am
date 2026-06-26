data "am_certificate" "my_certificate" {
  domain_key      = "example-domain"
  environment_id  = "DEFAULT"
  key             = "signing-cert"
  organization_id = "DEFAULT"
}