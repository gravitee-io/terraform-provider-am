resource "am_domain" "simple" {
  key           = "tf-smoke"
  name          = "[Terraform] Smoke Test Domain"
  description   = "Minimal runnable example for dev/QA smoke testing"
  path          = "/tf-smoke"
  data_plane_id = "default"
  enabled       = true
  # organization_id / environment_id default to "DEFAULT" (provider-level)
}
