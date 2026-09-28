# Compatibility

| Provider version  | Gravitee AM version |
|-------------------|---------------------|
| `< 1.0.0` (alpha) | 4.13.0              |

Provider releases are versioned independently of Gravitee Access Management.
Each published version also declares its compatible AM version on its
[Terraform Registry](https://registry.terraform.io/providers/gravitee-io/am/latest/docs)
overview page.

## Masked plugin configuration

AM masks sensitive values (passwords, client secrets, keystores) as `********` in the
`configuration` returned for identity providers, certificates and reporters. The provider
treats a masked value as equal to the configured one, so plans stay clean and the configured
secret is what Terraform keeps in state. A provider without this handling fails to apply those
resources against an AM version that masks. Because AM never returns the secret, Terraform
cannot detect a secret changed outside it; the next apply that changes the resource sends the
configured value again.
