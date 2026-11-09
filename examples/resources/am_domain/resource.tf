# A domain on the default data plane. Omitted settings blocks keep AM's defaults or inherit the
# organization's; a declared block with inherited = false overrides them.
resource "am_domain" "example" {
  key           = "example-domain"
  name          = "Example domain"
  description   = "An example authentication domain"
  path          = "/example-domain"
  data_plane_id = "default"

  login_settings = {
    inherited               = false
    forgot_password_enabled = true
    register_enabled        = true
  }

  password_settings = {
    inherited                  = false
    min_length                 = 12
    include_numbers            = true
    include_special_characters = true
    letters_in_mixed_case      = true
  }

  secret_expiration_settings = {
    enabled             = true
    expiry_time_seconds = 7776000 # 90 days
  }
}
