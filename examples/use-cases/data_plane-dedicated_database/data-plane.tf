# A data plane storing the runtime data (users, tokens...) of its domains in a database of their own, served
# by its own gateways.

resource "am_data_plane" "eu" {
  id          = "acme-eu"
  name        = "ACME EU"
  type        = "mongodb"
  gateway_url = "https://auth-eu.example.com"

  # Write-only: AM stores it but never returns it. Add username and password here when the database needs them.
  configuration = jsonencode({
    mongodb = {
      host   = var.mongodb_host
      port   = var.mongodb_port
      dbname = "acme-eu"
    }
  })
}

# Chosen on creation only: AM rejects an apply that moves the domain to another data plane.
resource "am_domain" "eu_customers" {
  key           = "acme-eu-customers"
  name          = "ACME EU customers"
  path          = "/acme-eu-customers"
  data_plane_id = am_data_plane.eu.id
}

# Read from the configuration: hosts and database name, never the credentials.
output "data_plane_hosts" {
  value = am_data_plane.eu.hosts
}

variable "mongodb_host" {
  type = string
}

variable "mongodb_port" {
  type    = number
  default = 27017
}
