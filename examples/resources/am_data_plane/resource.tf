resource "am_data_plane" "my_dataplane" {
  configuration   = "{ \"see\": \"documentation\" }"
  environment_id  = "DEFAULT"
  gateway_url     = "https://gateway-eu.example.com"
  id              = "acme-eu"
  name            = "ACME EU data plane"
  organization_id = "DEFAULT"
  type            = "mongodb"
}