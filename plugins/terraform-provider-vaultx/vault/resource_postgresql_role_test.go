package vault

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/hashicorp/vault/api"
)

func TestAccVaultPostgresqlRole_basic(t *testing.T) {
	var role api.Secret
	backendPath := fmt.Sprintf("pg-%s", acctest.RandString(5))
	name := fmt.Sprintf("role-%s", acctest.RandString(10))
	sql := `CREATE ROLE "{{name}}"`
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccVaultPostgresqlRoleConfig(backendPath, name, sql),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultPostgresqlRoleExists("vault_postgresql_role.foo", &role),
					testAccCheckVaultPostgresqlRoleAttributes(&role, sql),
				),
			},
			resource.TestStep{
				Config: testAccVaultPostgresqlSecretBackendConfig(backendPath),
				Check:  testAccCheckVaultPostgresqlRoleDestroy(backendPath, name),
			},
		},
	})
}

func TestAccVaultPostgresqlRole_disappears(t *testing.T) {
	var role api.Secret
	backendPath := fmt.Sprintf("pg-%s", acctest.RandString(5))
	name := fmt.Sprintf("role-%s", acctest.RandString(10))
	sql := `CREATE ROLE "{{name}}"`
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccVaultPostgresqlRoleConfig(backendPath, name, sql),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultPostgresqlRoleExists("vault_postgresql_role.foo", &role),
					testAccVaultPostgresqlRoleDisappear(backendPath, name),
				),
				ExpectNonEmptyPlan: true,
			},
			// Follow up w/ empty config should be empty, since the policy is gone.
			resource.TestStep{
				Config: "",
			},
		},
	})
}

func TestAccVaultPostgresqlRole_sqlDrift(t *testing.T) {
	var role api.Secret
	backendPath := fmt.Sprintf("pg-%s", acctest.RandString(5))
	name := fmt.Sprintf("role-%s", acctest.RandString(10))
	sql := `CREATE ROLE "{{name}}"`
	driftedSQL := `CREATE ROLE "{{name}}"; DROP TABLE bobby`

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccVaultPostgresqlRoleConfig(backendPath, name, sql),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultPostgresqlRoleExists("vault_postgresql_role.foo", &role),
					testAccVaultPostgresqlRoleDrift(backendPath, name, driftedSQL),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVaultPostgresqlRoleExists(
	key string, secret *api.Secret) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[key]
		client, err := testAccProvider.Meta().(ClientProvider).Client()
		if err != nil {
			return err
		}

		t, err := client.Logical().Read(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Error looking up secret: %s", err)
		}

		*secret = *t
		return nil
	}
}

func testAccCheckVaultPostgresqlRoleAttributes(
	secret *api.Secret,
	expectedSQL string,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return nil
	}
}

func testAccCheckVaultPostgresqlRoleDestroy(backendPath, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := testAccProvider.Meta().(ClientProvider).Client()
		if err != nil {
			return err
		}
		path := fmt.Sprintf("%s/roles/%s", backendPath, name)
		role, err := client.Logical().Read(path)
		if err != nil {
			return err
		}
		if role != nil {
			return fmt.Errorf("Role still exists: %s", path)
		}
		return nil
	}
}

func testAccVaultPostgresqlRoleDisappear(backendPath, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := testAccProvider.Meta().(ClientProvider).Client()
		if err != nil {
			return err
		}
		_, err = client.Logical().Delete(fmt.Sprintf("%s/roles/%s", backendPath, name))
		return err
	}
}

func testAccVaultPostgresqlRoleDrift(backendPath, name, sql string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := testAccProvider.Meta().(ClientProvider).Client()
		if err != nil {
			return err
		}
		path := fmt.Sprintf("%s/roles/%s", backendPath, name)
		_, err = client.Logical().Write(path, map[string]interface{}{"sql": sql})
		return err
	}
}

// Used to test destroy of the role w/o it being implicitly nuked by the
// backend unmounting.
func testAccVaultPostgresqlSecretBackendConfig(
	backendPath string) string {
	return fmt.Sprintf(`
resource "vault_secret_backend" "foo" {
	type = "postgresql"
  path = "%s"
	postgresql {
		connection_url    = "postgres://localhost/postgres?sslmode=disable"
		verify_connection = false
	}
}
`, backendPath)
}

func testAccVaultPostgresqlRoleConfig(
	backendPath, name, sql string) string {
	return fmt.Sprintf(`
resource "vault_secret_backend" "foo" {
	type = "postgresql"
  path = "%s"
	postgresql {
		connection_url    = "postgres://localhost/postgres?sslmode=disable"
		verify_connection = false
	}
}
resource "vault_postgresql_role" "foo" {
	backend = "${vault_secret_backend.foo.path}"
  name    = %q
  sql     = %q
}
`, backendPath, name, sql)
}
