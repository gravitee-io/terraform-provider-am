data "am_application" "my_application" {
  domain_key      = "example-domain"
  environment_id  = "DEFAULT"
  key             = "mobile-app"
  organization_id = "DEFAULT"
}
