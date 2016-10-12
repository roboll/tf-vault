package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/roboll/tf-vault/plugins/terraform-provider-vaultx/vault"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: vault.Provider,
	})
}
