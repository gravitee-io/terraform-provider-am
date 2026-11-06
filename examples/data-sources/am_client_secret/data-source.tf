data "am_client_secret" "my_clientsecret" {
  application_key = "mobile-app"
  domain_key      = "example-domain"
  environment_id  = "DEFAULT"
  key             = "ci-2026-10"
  organization_id = "DEFAULT"
}