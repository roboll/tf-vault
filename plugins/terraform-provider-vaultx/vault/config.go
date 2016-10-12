package vault

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"

	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/go-rootcerts"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/vault/api"
)

// ClientProvider is the interface consumed by resources.
type ClientProvider interface {
	// Client yields a Vault API client.
	Client() (*api.Client, error)
}

// Config is the builtin concrete implementation of VaultClientProvider
type Config struct {
	data *schema.ResourceData

	Address    string
	AuthToken  string
	CACertPool *x509.CertPool
	Insecure   bool
}

// Client implements VaultClientProvider
func (c *Config) Client() (*api.Client, error) {
	client, err := api.NewClient(c.VaultConfig())
	if err != nil {
		return nil, err
	}
	client.SetToken(c.AuthToken)
	return client, nil
}

// VaultConfig prepares a Vault API config ready to be passed into a client.
func (c *Config) VaultConfig() *api.Config {
	vc := api.DefaultConfig()
	vc.Address = c.Address
	tlsConfig := vc.HttpClient.Transport.(*http.Transport).TLSClientConfig
	tlsConfig.InsecureSkipVerify = c.Insecure
	if c.CACertPool != nil {
		tlsConfig.RootCAs = c.CACertPool
	}
	return vc
}

// SetTokenForClientCert loads provided client cert / key and uses them to
// fetch and store an AuthToken on this Config
func (c *Config) SetTokenForClientCert(cert, key, mount string) error {
	// Empty cert/key is noop
	if cert == "" && key == "" {
		return nil
	}
	if cert == "" || key == "" {
		return fmt.Errorf("Both client_cert and client_key must be provided together.")
	}
	clientCert, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return err
	}

	// Prepare a Vault API client using the loaded client certs
	vc := c.VaultConfig()
	tlsConfig := vc.HttpClient.Transport.(*http.Transport).TLSClientConfig
	tlsConfig.Certificates = []tls.Certificate{clientCert}
	client, err := api.NewClient(vc)
	if err != nil {
		return errwrap.Wrapf("Error setting up Vault client for client cert auth: {{err}}", err)
	}

	if mount == "" {
		mount = "cert"
	}

	path := fmt.Sprintf("auth/%s/login", mount)
	secret, err := client.Logical().Write(path, nil)
	if err != nil {
		return err
	}
	if secret == nil {
		return fmt.Errorf("empty response from credential provider")
	}
	c.AuthToken = secret.Auth.ClientToken
	return nil
}

// SetCACertPool configures the CACertPool by loading the cert or path
// specified.
func (c *Config) SetCACertPool(caCert, caPath string) error {
	if caPath != "" {
		certPool, err := rootcerts.LoadCAPath(caPath)
		if err != nil {
			return errwrap.Wrapf("Error loading CA path: {{err}}", err)
		}
		c.CACertPool = certPool
	}
	if caCert != "" {
		certPool, err := rootcerts.LoadCAFile(caCert)
		if err != nil {
			return errwrap.Wrapf("Error loading CA cert: {{err}}", err)
		}
		c.CACertPool = certPool
	}
	if caPath == "" && caCert == "" {
		log.Println("using system cas roboll")
		certPool, err := rootcerts.LoadSystemCAs()
		if err != nil {
			return errwrap.Wrapf("Error loading System CA certs: {{err}}", err)
		}
		c.CACertPool = certPool
	}
	return nil
}
