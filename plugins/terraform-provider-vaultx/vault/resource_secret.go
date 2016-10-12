package vault

import (
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceVaultSecret() *schema.Resource {
	return &schema.Resource{
		Create: resourceVaultSecretCreate,
		// Yay for PUT
		Update: resourceVaultSecretCreate,
		Exists: resourceVaultSecretExists,
		Read:   resourceVaultSecretRead,
		Delete: resourceVaultSecretDelete,

		Schema: map[string]*schema.Schema{
			"path": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"ttl": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"data": &schema.Schema{
				Type:      schema.TypeMap,
				Optional:  true,
				Sensitive: true,
			},

			"token": &schema.Schema{
				Type:        schema.TypeMap,
				Optional:    true,
				Sensitive:   true,
				Description: "Overrides the provider's token for authentication when interacting with this secret. Helpful for interacting with a cubbyhole backend.",
			},

			"ignore_delete": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Skips deleting the secret. Useful for configuration values that do not allow delete.",
			},
			"ignore_read": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Skips reading the secret. Useful for configuration values that do not allow read.",
			},
		},
	}
}

func resourceVaultSecretCreate(d *schema.ResourceData, meta interface{}) error {
	client, err := meta.(ClientProvider).Client()
	if err != nil {
		return err
	}

	data := d.Get("data").(map[string]interface{})
	if ttl := d.Get("ttl").(string); ttl != "" {
		data["ttl"] = ttl
	}

	written, err := client.Logical().Write(d.Get("path").(string), data)
	if err != nil {
		return err
	}

	if written != nil {
		for key, val := range written.Data {
			data[key] = val
		}
		d.Set("data", data)
	}

	d.SetId(d.Get("path").(string))
	return nil
}

func resourceVaultSecretExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	client, err := meta.(ClientProvider).Client()
	if err != nil {
		return false, err
	}

	if d.Get("ignore_read").(bool) {
		return true, nil
	}
	secret, err := client.Logical().Read(d.Get("path").(string))
	if err != nil {
		return false, err
	}

	exists := (secret != nil)
	return exists, nil
}

func resourceVaultSecretRead(d *schema.ResourceData, meta interface{}) error {
	client, err := meta.(ClientProvider).Client()
	if err != nil {
		return err
	}

	if d.Get("ignore_read").(bool) {
		return nil
	}
	secret, err := client.Logical().Read(d.Get("path").(string))
	if err != nil {
		return err
	}

	if ttl, ok := secret.Data["ttl"]; ok {
		d.Set("ttl", ttl.(string))
	}

	delete(secret.Data, "ttl")

	if err := d.Set("data", secret.Data); err != nil {
		return err
	}

	return nil
}

func isSecretNotFoundError(err error) bool {
	return strings.Contains(err.Error(), "bad token")
}

func resourceVaultSecretDelete(d *schema.ResourceData, meta interface{}) error {
	client, err := meta.(ClientProvider).Client()
	if err != nil {
		return err
	}

	if !d.Get("ignore_delete").(bool) {
		_, err = client.Logical().Delete(d.Get("path").(string))
		if err != nil {
			return err
		}
	}

	d.SetId("")
	return nil
}
