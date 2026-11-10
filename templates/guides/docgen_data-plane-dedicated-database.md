---
page_title: "Domains on a dedicated data plane"
subcategory: "Data planes"
---

# Domains on a dedicated data plane

A data plane holds the runtime data (users, tokens, consents...) of the domains bound to it. This guide
creates a data plane on a MongoDB database of its own, for example to keep the users of a region apart, and a
domain bound to it.

## Before you start

* The provider reaches the AM Automation API with `AM_SERVER_URL` and a service-account token in `AM_SA_TOKEN`.
* A MongoDB the AM management API and the gateways reach: its host in `TF_VAR_mongodb_host`, and the user AM
  connects with in `TF_VAR_mongodb_username` and `TF_VAR_mongodb_password`, authenticated against the `admin`
  database (`authSource`).
* Gateways bound to this data plane: started with `repositories.gateway.dataPlane.id: acme-eu` in their
  `gravitee.yml` (or `GRAVITEE_REPOSITORIES_GATEWAY_DATAPLANE_ID=acme-eu`), and served at
  `https://auth-eu.example.com`. `gateway_url` only tells AM where they are: AM does not deploy them.

## Configuration

```terraform
# A data plane storing the runtime data (users, tokens...) of its domains in a database of their own, served
# by its own gateways.

resource "am_data_plane" "eu" {
  id          = "acme-eu"
  name        = "ACME EU"
  type        = "mongodb"
  gateway_url = "https://auth-eu.example.com"

  # Write-only: AM stores it but never returns it. The fields are those of the dataplane-am-mongodb plugin.
  configuration = jsonencode({
    mongodb = {
      host       = var.mongodb_host
      port       = var.mongodb_port
      dbname     = "acme-eu"
      username   = var.mongodb_username
      password   = var.mongodb_password
      authSource = "admin"
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

# Left unset (null) for a MongoDB without authentication.
variable "mongodb_username" {
  type    = string
  default = null
}

variable "mongodb_password" {
  type      = string
  sensitive = true
  default   = null
}

```

## How it differs from the console and the Management API

* `configuration` is write-only: AM never returns it, credentials or not. Terraform compares it with the
  previous apply, not with AM: a change made in the console does not show in the plan, and the next change of
  the configuration in Terraform sends it back as declared.
* After an import, the state has no configuration: the first plan shows it as changed, and the apply sends it.
* `hosts` and `database` are read back from AM, from the configuration it stores: use them to check what AM
  connects to, since a misspelled field of `configuration` is not reported. The credentials are never read back.
* `type` cannot change once created: AM rejects the apply.
* A domain chooses its data plane on creation only: an apply that changes `data_plane_id` is rejected. To move
  users to another data plane, create a new domain. A domain created without `data_plane_id` gets one of the
  environment's data planes, chosen by AM.
