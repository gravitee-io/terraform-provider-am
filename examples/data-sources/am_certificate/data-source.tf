data "am_certificate" "my_certificate" {
  domain_key = "customers"
  env_id     = "DEFAULT"
  key        = "signing-cert"
  org_id     = "DEFAULT"
}