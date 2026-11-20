# The certificate signing the tokens: the domain and the applications follow it.
locals {
  # The domain names it by key: the certificate lives in the domain, so the domain cannot wait for it.
  signing_certificate_key = "signing-2027"
  # The applications reference it, so Terraform deletes them before the certificate.
  signing_certificate = am_certificate.signing_2027.key
}
