resource "am_domain" "my_domain" {
  account_settings = {
    account_blocked_duration                         = 7200
    auto_login_after_registration                    = false
    auto_login_after_reset_password                  = false
    complete_registration_when_reset_password        = false
    default_identity_provider_for_registration       = "users-idp"
    delete_passwordless_devices_after_reset_password = false
    dynamic_user_registration                        = false
    inherited                                        = true
    login_attempts_detection_enabled                 = false
    login_attempts_reset_time                        = 600
    max_login_attempts                               = 10
    mfa_challenge_attempts_detection_enabled         = false
    mfa_challenge_attempts_reset_time                = 600
    mfa_challenge_max_attempts                       = 3
    mfa_challenge_send_verify_alert_email            = false
    redirect_uri_after_registration                  = "https://app.example.com/welcome"
    redirect_uri_after_reset_password                = "https://app.example.com/signin"
    remember_me                                      = false
    remember_me_duration                             = 604800
    reset_password_confirm_identity                  = false
    reset_password_custom_form                       = false
    reset_password_custom_form_fields = [
      {
        key   = "email"
        label = "Email"
        type  = "email"
      }
    ]
    reset_password_invalidate_tokens       = false
    send_recover_account_email             = false
    send_verify_registration_account_email = false
  }
  certificate_settings = {
    fallback_certificate = "default"
  }
  cors_settings = {
    allow_credentials = false
    allowed_headers = [
      "Authorization",
      "Content-Type",
    ]
    allowed_methods = [
      "GET",
      "POST",
      "PUT",
      "DELETE",
    ]
    allowed_origins = [
      "https://app.example.com",
    ]
    enabled = false
    max_age = 86400
  }
  data_plane_id  = "default"
  description    = "Customer-facing authentication domain"
  enabled        = true
  environment_id = "DEFAULT"
  key            = "customers"
  login_settings = {
    certificate_based_auth_enabled        = false
    certificate_based_auth_url            = "...my_certificate_based_auth_url..."
    enforce_password_policy_enabled       = true
    forgot_password_enabled               = false
    hide_form                             = false
    identifier_first_enabled              = false
    inherited                             = true
    magic_link_auth_enabled               = false
    passwordless_device_naming_enabled    = false
    passwordless_enabled                  = false
    passwordless_enforce_password_enabled = false
    passwordless_enforce_password_max_age = 6
    passwordless_remember_device_enabled  = false
    register_enabled                      = false
    remember_me_enabled                   = false
    reset_password_on_expiration          = true
  }
  name = "Customers"
  oidc = {
    ciba_settings = {
      auth_req_expiry        = 600
      binding_message_length = 256
      enabled                = false
      token_req_interval     = 5
    }
    client_registration_settings = {
      allow_http_scheme_redirect_uri                = false
      allow_localhost_redirect_uri                  = false
      allow_redirect_uri_params_expression_language = false
      allow_wild_card_redirect_uri                  = false
      allowed_scopes = [
        "..."
      ]
      allowed_scopes_enabled  = true
      client_template_enabled = true
      default_scopes = [
        "..."
      ]
      dynamic_client_registration_enabled      = true
      open_dynamic_client_registration_enabled = true
    }
    post_logout_redirect_uris = [
      "..."
    ]
    redirect_uri_strict_matching = false
    request_uris = [
      "..."
    ]
    security_profile_settings = {
      enable_fapi_brazil = false
      enable_plain_fapi  = false
    }
  }
  organization_id = "DEFAULT"
  password_settings = {
    exclude_passwords_in_dictionary       = true
    exclude_user_profile_info_in_password = true
    expiry_duration                       = 6
    include_numbers                       = true
    include_special_characters            = true
    inherited                             = true
    letters_in_mixed_case                 = false
    max_consecutive_letters               = 7
    max_length                            = 128
    min_length                            = 8
    old_passwords                         = 6
    password_history_enabled              = false
  }
  path = "/customers"
  saml = {
    certificate = "signing-cert"
    enabled     = false
    entity_id   = "https://auth.example.com/saml2/idp/entity"
  }
  scim = {
    enabled               = false
    idp_selection_enabled = false
    idp_selection_rule    = "...my_idp_selection_rule..."
  }
  secret_expiration_settings = {
    enabled             = false
    expiry_time_seconds = 7776000
  }
  self_service_account_management_settings = {
    enabled = false
    reset_password = {
      old_password_required = false
      token_age             = 2
    }
  }
  tags = [
    "eu",
    "production",
  ]
  token_exchange_settings = {
    allow_delegation    = false
    allow_impersonation = true
    allowed_actor_token_types = [
      "urn:ietf:params:oauth:token-type:access_token",
      "urn:ietf:params:oauth:token-type:id_token",
    ]
    allowed_requested_token_types = [
      "urn:ietf:params:oauth:token-type:access_token",
      "urn:ietf:params:oauth:token-type:id_token",
    ]
    allowed_subject_token_types = [
      "urn:ietf:params:oauth:token-type:access_token",
      "urn:ietf:params:oauth:token-type:id_token",
    ]
    enabled              = false
    max_delegation_depth = 25
    token_exchange_o_auth_settings = {
      inherited      = true
      scope_handling = "DOWNSCOPING"
    }
    trusted_issuers = [
      {
        certificate           = "...my_certificate..."
        issuer                = "https://issuer.example.com"
        jwks_uri              = "https://issuer.example.com/.well-known/jwks.json"
        key_resolution_method = "PEM"
        scope_mappings = {
          key = "value"
        }
        user_binding_criteria = [
          {
            attribute  = "emails.value"
            expression = "{#token['email']}"
          }
        ]
        user_binding_enabled = false
      }
    ]
  }
  uma = {
    enabled = false
  }
  vhosts = [
    {
      host                = "auth.example.com"
      override_entrypoint = false
      path                = "/customers"
    }
  ]
  web_authn_settings = {
    attestation_conveyance_preference = "NONE"
    authenticator_attachment          = "CROSS_PLATFORM"
    certificates = {
      key = {
        # ...
      }
    }
    enforce_authenticator_integrity         = false
    enforce_authenticator_integrity_max_age = 10
    force_registration                      = false
    origin                                  = "https://auth.example.com"
    relying_party_id                        = "auth.example.com"
    relying_party_name                      = "Example Inc."
    require_resident_key                    = false
    user_verification                       = "PREFERRED"
  }
}