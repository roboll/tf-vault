package main

import (
	"testing"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/roboll/tf-vault/plugins/terraform-provider-vaultx/vault"
)

func TestProvider(t *testing.T) {
	if err := vault.Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}
