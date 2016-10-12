package vault

import "github.com/hashicorp/terraform/helper/schema"

func dataVaultSecret() *schema.Resource {
	return &schema.Resource{
		Read: dataVaultSecretRead,

		Schema: map[string]*schema.Schema{
			"path": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"data": &schema.Schema{
				Type:      schema.TypeMap,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func dataVaultSecretRead(d *schema.ResourceData, meta interface{}) error {
	client, err := meta.(ClientProvider).Client()
	if err != nil {
		return err
	}

	secret, err := client.Logical().Read(d.Get("path").(string))
	if err != nil {
		return err
	}

	d.SetId(d.Get("path").(string))
	if err := d.Set("data", secret.Data); err != nil {
		return err
	}

	return nil
}
