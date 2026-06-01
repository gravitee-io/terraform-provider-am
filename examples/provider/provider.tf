terraform {
  required_providers {
    am = {
      source  = "gravitee-io/am"
      version = "0.0.1"
    }
  }
}

provider "am" {
  server_url = "..." # Optional - can use AM_SERVER_URL environment variable
}