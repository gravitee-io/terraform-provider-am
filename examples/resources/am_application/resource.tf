resource "am_application" "my_application" {
  certificate    = "signing-cert"
  client_id      = "mobile-app-client"
  description    = "Customer mobile application"
  domain_key     = "example-domain"
  enabled        = true
  environment_id = "DEFAULT"
  identity_providers = [
    {
      identity       = "customer-users"
      priority       = 0
      selection_rule = "{#request.params['idp'][0] == 'customers'}"
    }
  ]
  key  = "mobile-app"
  kind = "USER_EMBEDDED"
  metadata = {
    key = jsonencode("value")
  }
  name            = "Mobile app"
  organization_id = "DEFAULT"
  settings = {
    account = {
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
    advanced = {
      flows_inherited = true
      skip_consent    = false
    }
    cookie_settings = {
      inherited = true
      session = {
        persistent = false
      }
    }
    login = {
      certificate_based_auth_enabled        = false
      certificate_based_auth_url            = "...my_certificate_based_auth_url..."
      forgot_password_enabled               = false
      hide_form                             = false
      identifier_first_enabled              = false
      inherited                             = true
      magic_link_auth_enabled               = false
      passwordless_device_naming_enabled    = false
      passwordless_enabled                  = false
      passwordless_enforce_password_enabled = false
      passwordless_enforce_password_max_age = 10
      passwordless_remember_device_enabled  = false
      register_enabled                      = false
      remember_me_enabled                   = false
      reset_password_on_expiration          = false
    }
    mfa = {
      adaptive_authentication_rule = "...my_adaptive_authentication_rule..."
      challenge = {
        active         = false
        challenge_rule = "{#context.attributes['geoip']['country_iso_code'] == 'FR'}"
        type           = "required"
      }
      enroll = {
        active                 = false
        enrollment_rule        = "{#context.attributes['user'].roles.contains('guest')}"
        enrollment_skip_active = false
        enrollment_skip_rule   = "...my_enrollment_skip_rule..."
        force_enrollment       = false
        skip_time_seconds      = 36000
        type                   = "required"
      }
      enrollment = {
        force_enrollment  = false
        skip_time_seconds = 36000
      }
      factor = {
        application_factors = [
          {
            id             = "id:5e8f1c0a-2b3d-4f5e-8a9b-0c1d2e3f4a5b"
            selection_rule = "{#context.attributes['user'].roles.contains('admin')}"
          }
        ]
        default_factor_id = "id:5e8f1c0a-2b3d-4f5e-8a9b-0c1d2e3f4a5b"
      }
      login_rule = "...my_login_rule..."
      remember_device = {
        active                  = false
        device_identifier_id    = "id:5e8f1c0a-2b3d-4f5e-8a9b-0c1d2e3f4a5b"
        expiration_time_seconds = 36000
        skip_remember_device    = false
      }
      step_up_authentication = {
        active                      = false
        step_up_authentication_rule = "{#request.params['scope'][0].contains('write')}"
      }
      step_up_authentication_rule = "...my_step_up_authentication_rule..."
    }
    oauth = {
      access_token_validity_seconds        = 7200
      authorization_encrypted_response_alg = "...my_authorization_encrypted_response_alg..."
      authorization_encrypted_response_enc = "...my_authorization_encrypted_response_enc..."
      authorization_signed_response_alg    = "...my_authorization_signed_response_alg..."
      client_uri                           = "...my_client_uri..."
      contacts = [
        "ops@example.com",
      ]
      cross_app_access_settings = {
        enabled = false
        resource_servers = [
          {
            client_id          = "acme-calendar-client"
            resource_server_id = "id:7a1b2c3d-4e5f-4a6b-8c7d-9e0f1a2b3c4d"
            trust_domain_id    = "id:5e8f1c0a-2b3d-4f5e-8a9b-0c1d2e3f4a5b"
          }
        ]
      }
      default_ac_rvalues = [
        "..."
      ]
      default_max_age                      = 3600
      disable_refresh_token_rotation       = false
      dpop_bound_access_tokens             = false
      enhance_scopes_with_user_permissions = false
      force_pkce                           = false
      force_s256_code_challenge_method     = false
      grant_types = [
        "authorization_code",
        "refresh_token",
      ]
      id_jag_validity_seconds         = 300
      id_token_encrypted_response_alg = "...my_id_token_encrypted_response_alg..."
      id_token_encrypted_response_enc = "...my_id_token_encrypted_response_enc..."
      id_token_signed_response_alg    = "...my_id_token_signed_response_alg..."
      id_token_validity_seconds       = 14400
      initiate_login_uri              = "https://app.example.com/login"
      jwks = {
        key = jsonencode("value")
      }
      jwks_uri = "https://app.example.com/.well-known/jwks.json"
      lightweight_jwt_settings = {
        allowed_access_token_claims = [
          "..."
        ]
        allowed_id_token_claims = [
          "..."
        ]
        enabled = false
      }
      logo_uri               = "...my_logo_uri..."
      opt_in_scope_selection = true
      policy_uri             = "...my_policy_uri..."
      post_logout_redirect_uris = [
        "https://app.example.com/logged-out",
      ]
      redirect_uris = [
        "https://app.example.com/callback",
      ]
      refresh_token_validity_seconds = 14400
      request_object_encryption_alg  = "...my_request_object_encryption_alg..."
      request_object_encryption_enc  = "...my_request_object_encryption_enc..."
      request_object_signing_alg     = "...my_request_object_signing_alg..."
      request_uris = [
        "..."
      ]
      require_auth_time = false
      response_types = [
        "code",
      ]
      scope_settings = [
        {
          default_scope  = false
          required_scope = false
          scope          = "openid"
          scope_approval = 86400
        }
      ]
      sector_identifier_uri                      = "...my_sector_identifier_uri..."
      silent_re_authentication                   = false
      single_sign_out                            = false
      subject_type                               = "public"
      tls_client_auth_san_dns                    = "...my_tls_client_auth_san_dns..."
      tls_client_auth_san_email                  = "...my_tls_client_auth_san_email..."
      tls_client_auth_san_ip                     = "...my_tls_client_auth_san_ip..."
      tls_client_auth_san_uri                    = "...my_tls_client_auth_san_uri..."
      tls_client_auth_subject_dn                 = "CN=app.example.com"
      tls_client_certificate_bound_access_tokens = false
      token_custom_claims = [
        {
          claim_name  = "department"
          claim_value = "{#context.attributes['user']['additionalInformation']['department']}"
          token_type  = "access_token"
        }
      ]
      token_endpoint_auth_method      = "client_secret_basic"
      token_endpoint_auth_signing_alg = "RS256"
      token_exchange_o_auth_settings = {
        inherited      = true
        scope_handling = "downscoping"
      }
      tos_uri = "...my_tos_uri..."
      userinfo_custom_claims = [
        {
          claim_name  = "department"
          claim_value = "{#context.attributes['user']['additionalInformation']['department']}"
        }
      ]
      userinfo_encrypted_response_alg = "...my_userinfo_encrypted_response_alg..."
      userinfo_encrypted_response_enc = "...my_userinfo_encrypted_response_enc..."
      userinfo_signed_response_alg    = "...my_userinfo_signed_response_alg..."
    }
    risk_assessment = {
      device_assessment = {
        enabled = false
        thresholds = {
          key = 1.23
        }
      }
      enabled = false
      geo_velocity_assessment = {
        enabled = false
        thresholds = {
          key = 1.23
        }
      }
      ip_reputation_assessment = {
        enabled = false
        thresholds = {
          key = 1.23
        }
      }
    }
    saml = {
      assertion_attributes = [
        {
          name  = "email"
          value = "{#context.attributes['user'].email}"
        }
      ]
      assertion_validity_seconds    = 300
      attribute_consume_service_url = "https://app.example.com/saml/acs"
      audiences = [
        "..."
      ]
      certificate                        = "...my_certificate..."
      data_encryption_algorithm          = "http://www.w3.org/2009/xmlenc11#aes256-gcm"
      entity_id                          = "https://app.example.com"
      include_assertion_conditions       = false
      key_transport_encryption_algorithm = "http://www.w3.org/2009/xmlenc11#rsa-oaep"
      name_id_mapping                    = "{#context.attributes['user'].username}"
      not_before_time_skew_seconds       = 30
      not_on_or_after_time_skew_seconds  = 30
      response_binding                   = "urn:oasis:names:tc:SAML:2.0:bindings:custom:Initial-Request"
      single_logout_service_url          = "https://app.example.com/saml/logout"
      want_assertions_encrypted          = false
      want_assertions_signed             = false
      want_response_signed               = true
    }
    secret_expiration_settings = {
      enabled             = false
      expiry_time_seconds = 7776000
    }
    workload_identity_settings = {
      subject            = "spiffe://prod.example.org/ns/payments/sa/api"
      subject_match_mode = "exact"
      trust_domain       = "prod.example.org"
    }
  }
  template = false
  type     = "native"
}