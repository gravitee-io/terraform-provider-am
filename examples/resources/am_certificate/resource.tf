resource "am_certificate" "my_certificate" {
  configuration = "{\"jks\":{\"content\":\"...\",\"name\":\"keystore.jks\"},\"storepass\":\"secret\",\"alias\":\"mykey\",\"keypass\":\"secret\"}"
  domain_key    = "customers"
  env_id        = "DEFAULT"
  key           = "signing-cert"
  name          = "Signing certificate"
  org_id        = "DEFAULT"
  system        = false
  type          = "javakeystore-am-certificate"
}