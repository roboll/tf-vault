package vault

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceVaultPostgresqlRole() *schema.Resource {
	return &schema.Resource{
		Create: resourceVaultPostgresqlRoleCreate,
		Update: resourceVaultPostgresqlRoleCreate,
		Read:   resourceVaultPostgresqlRoleRead,
		Delete: resourceVaultPostgresqlRoleDelete,
		Exists: resourceVaultPostgresqlRoleExists,

		Schema: map[string]*schema.Schema{
			"backend": &schema.Schema{
				Description: "The path to the PostgreSQL backend in which to create this role",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "postgresql",
			},

			"name": &schema.Schema{
				Description: "The name of this role. Must be unique per backend.",
				Type:        schema.TypeString,
				Required:    true,
			},

			"sql": &schema.Schema{
				Description: strings.TrimSpace(`
					The SQL statements executed to create and configure the role. Must be semi-colon separated. The '{{name}}', '{{password}}' and '{{expiration}}' values will be substituted.
					`),
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceVaultPostgresqlRoleCreate(d *schema.ResourceData, meta interface{}) error {
	client, err := meta.(ClientProvider).Client()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("%s/roles/%s",
		d.Get("backend").(string), d.Get("name").(string))

	data := map[string]interface{}{
		"sql": d.Get("sql").(string),
	}
	_, err = client.Logical().Write(path, data)
	if err != nil {
		return err
	}

	d.SetId(path)
	return nil
}

func resourceVaultPostgresqlRoleExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	client, err := meta.(ClientProvider).Client()
	if err != nil {
		return false, err
	}
	path := fmt.Sprintf("%s/roles/%s",
		d.Get("backend").(string), d.Get("name").(string))

	secret, err := client.Logical().Read(path)
	if err != nil {
		return false, err
	}

	exists := secret != nil
	return exists, nil
}

func resourceVaultPostgresqlRoleRead(d *schema.ResourceData, meta interface{}) error {
	client, err := meta.(ClientProvider).Client()
	if err != nil {
		return err
	}

	role, err := client.Logical().Read(d.Id())
	if err != nil {
		return err
	}
	if role == nil || role.Data == nil {
		return fmt.Errorf("Got unexpected nil role for path %s", d.Id())
	}

	sql := ""
	if v, ok := role.Data["sql"]; ok {
		sql = v.(string)
	}
	d.Set("sql", sql)

	return nil
}

func resourceVaultPostgresqlRoleDelete(d *schema.ResourceData, meta interface{}) error {
	client, err := meta.(ClientProvider).Client()
	if err != nil {
		return err
	}

	_, err = client.Logical().Delete(d.Id())
	if err != nil {
		return err
	}

	return nil
}
