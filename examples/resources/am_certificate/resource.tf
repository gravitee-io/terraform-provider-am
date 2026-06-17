resource "am_certificate" "my_certificate" {
  configuration   = "{\"jks\":{\"content\":\"...\",\"name\":\"keystore.jks\"},\"storepass\":\"secret\",\"alias\":\"mykey\",\"keypass\":\"secret\"}"
  domain_key      = "example-domain"
  environment_id  = "DEFAULT"
  key             = "signing-cert"
  name            = "Signing certificate"
  organization_id = "DEFAULT"
  system          = false
  type            = "javakeystore-am-certificate"
}