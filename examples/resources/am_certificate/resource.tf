resource "am_certificate" "my_certificate" {
  configuration   = "{\"jks\":{\"content\":\"...\",\"name\":\"keystore.jks\"},\"storepass\":\"secret\",\"alias\":\"mykey\",\"keypass\":\"secret\"}"
  domain_key      = "customers"
  environment_id  = "...my_environment_id..."
  key             = "signing-cert"
  name            = "Signing certificate"
  organization_id = "...my_organization_id..."
  system          = false
  type            = "javakeystore-am-certificate"
}