# tf-vault [![CircleCI](https://circleci.com/gh/roboll/tf-vault.svg?style=svg)](https://circleci.com/gh/roboll/tf-vault)

## Deployment

Terraform module for deploying [Vault](https://vaultproject.io) on [CoreOS](https://coreos.com) with [rkt](https://github.com/coreos/rkt).

By default, uses https://quay.io/roboll/vault to run vault, and https://quay.io/roboll/prom-statsd-exporter to run the prometheus statsd exporter. Vault sends statsd style metrics to the exporter, which exposes them for prometheus scraping on port 9102.

Vault runs in HA mode via non-secured `etcd` running on all nodes. Secrets are stored in s3.

Certificates are provisioned from letsencrypt using a dns challenge on route53. To distribute new certificates, machines must be restarted - manual orchestration is necessary to ensure no downtime. Certificates are shipped to hosts encrypted using kms, and are decrypted on the host.

Vault instances are fronted by an ELB, as well as addressable by their own dns for direct unseal operations.

## Plugins

Terraform providers for [Vault](https://vaultproject.io). There are two plugins here.

The first one packages the code [here](https://github.com/hashicorp/terraform/tree/f-vault/builtin/providers/vault) with some modifications, and packaged as a standalone provider. It's called vault**x** so that when/if that code gets incorporated to mainline Terraform, there won't be ambiguity between which version is in use for a given resource. [docs](https://github.com/hashicorp/terraform/tree/f-vault/website/source/docs/providers/vault)

The second is a plugin for initializing new vault clusters. It doesn't follow the terraform ways - it does nothing on destroy, and on apply it initializes and unseals the cluster, writing the keys out to a local file. **Use at your own risk.**

### Usage

```
resource vaultinit_init vault {
  instances = ["10.0.1.1:8200", "10.0.1.2:8200"]   // vault instances
  address = ""                                     // hostname to validate ssl and wait for response

  secret_shares = 5
  secret_threshold = 3
  pgp_keys = []

  wait_for_response = true                         // wait for instances to appear healthy

  unseal_keys_output_file = "${path.root}/vault-keys.txt"
}

output root_token { value = "${vaultinit_init.vault.root_token}" }
```

### Get It

`go get github.com/roboll/terraform-vault/plugins/...`

_or_

`curl -L -o /usr/local/bin/terraform-provider-vaultinit https://github.com/roboll/terraform-vault/releases/download/{VERSION}/terraform-provider-vaultinit{OS}_{ARCH}`
`curl -L -o /usr/local/bin/terraform-provider-vaultx https://github.com/roboll/terraform-vault/releases/download/{VERSION}/terraform-provider-vaultx{OS}_{ARCH}`
