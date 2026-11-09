# A data plane on its own MongoDB database. The configuration is write-only: AM never returns it.
resource "am_data_plane" "eu" {
  id          = "acme-eu"
  name        = "ACME EU data plane"
  type        = "mongodb"
  gateway_url = "https://gateway-eu.example.com"
  configuration = jsonencode({
    mongodb = {
      host   = "mongodb"
      port   = 27017
      dbname = "gravitee-am-eu"
    }
  })
}
