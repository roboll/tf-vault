package vault

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

// Provider returns a schema.Provider for managing Packet infrastructure.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"address": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VAULT_ADDR", ""),
			},

			"ca_cert": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VAULT_CACERT", ""),
			},

			"ca_path": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VAULT_CAPATH", ""),
			},

			"token": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("VAULT_TOKEN", ""),
			},

			"client_cert": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VAULT_CLIENT_CERT", ""),
			},

			"client_key": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VAULT_CLIENT_KEY", ""),
			},

			"auth_config": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				Default:  map[string]interface{}{},
			},

			"allow_unverified_ssl": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VAULT_SKIP_VERIFY", false),
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"vaultx_audit_backend":   resourceVaultAuditBackend(),
			"vaultx_auth_backend":    resourceVaultAuthBackend(),
			"vaultx_secret_backend":  resourceVaultSecretBackend(),
			"vaultx_policy":          resourceVaultPolicy(),
			"vaultx_postgresql_role": resourceVaultPostgresqlRole(),
			"vaultx_secret":          resourceVaultSecret(),
			"vaultx_token":           resourceVaultToken(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"vaultx_secret": dataVaultSecret(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := &Config{
		Address:   d.Get("address").(string),
		Insecure:  d.Get("allow_unverified_ssl").(bool),
		AuthToken: d.Get("token").(string),
	}

	caCert := d.Get("ca_cert").(string)
	caPath := d.Get("ca_path").(string)
	if err := config.SetCACertPool(caCert, caPath); err != nil {
		return nil, err
	}

	var authConfig map[string]string
	for k, v := range d.Get("auth_config").(map[string]interface{}) {
		authConfig[k] = v.(string)
	}

	cert := d.Get("client_cert").(string)
	key := d.Get("client_key").(string)
	mount := authConfig["mount"]
	if err := config.SetTokenForClientCert(cert, key, mount); err != nil {
		return nil, err
	}

	return config, nil
}
