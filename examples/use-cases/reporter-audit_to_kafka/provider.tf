terraform {
  required_providers {
    am = {
      source = "gravitee-io/am"
    }
  }
}

provider "am" {
  # Can be set/overridden with the AM_SERVER_URL environment variable
  #server_url = "http://localhost:8093/automation"

  # Can be set/overridden with the AM_SA_TOKEN environment variable
  #bearer_auth = "xxx" # a JWT or an opaque service-account access token

  # organization_id and environment_id default to "DEFAULT".
  # Can be set/overridden with the AM_ORG_ID / AM_ENV_ID environment variables.
  #organization_id = "DEFAULT"
  #environment_id  = "DEFAULT"
}
