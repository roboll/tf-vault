package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/vault/api"
	"github.com/satori/go.uuid"
)

func resourceVaultInitUnseal() *schema.Resource {
	return &schema.Resource{
		Create: resourceInitCreate,
		Read:   noop,
		Update: noop,
		Delete: resourceInitDelete,

		Schema: map[string]*schema.Schema{
			"instances": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
				ForceNew: false,
			},

			"allow_unverified_ssl": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VAULT_SKIP_VERIFY", false),
				ForceNew:    false,
			},

			"response_timeout": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "5m",
				ForceNew: false,
			},

			"address": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},

			"wait_for_ready": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				ForceNew: false,
			},

			"unseal_keys_output_file": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},

			"secret_shares": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: false,
				Default:  5,
			},

			"secret_threshold": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: false,
				Default:  3,
			},

			"pgp_keys": &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				ForceNew: false,
			},

			"root_token": &schema.Schema{
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceInitCreate(d *schema.ResourceData, meta interface{}) error {
	vc := api.DefaultConfig()
	tlsConfig := vc.HttpClient.Transport.(*http.Transport).TLSClientConfig
	tlsConfig.InsecureSkipVerify = d.Get("allow_unverified_ssl").(bool)

	instances := []string{}
	il := d.Get("instances").([]interface{})
	for _, instance := range il {
		instances = append(instances, instance.(string))
	}
	vc.Address = instances[0]

	client, err := api.NewClient(vc)
	if err != nil {
		return err
	}

	init := make(chan interface{})
	go func() {
		exit := make(chan struct{})
		waitString := d.Get("response_timeout").(string)
		wait, err := time.ParseDuration(waitString)
		if err != nil {
			init <- err
			return
		}
		time.AfterFunc(wait, func() {
			exit <- struct{}{}
		})
		var lasterr error
		for {
			select {
			case <-exit:
				if lasterr == nil {
					lasterr = errors.New("unexpected error")
				}
				init <- lasterr
				return
			default:
				initialized, err := client.Sys().InitStatus()
				if err != nil {
					lasterr = err
				} else {
					init <- initialized
					return
				}
			}
		}
	}()

	result := <-init
	switch res := result.(type) {
	case error:
		return fmt.Errorf("error: failed to check vault status: %s", res)
	case bool:
		if res {
			return fmt.Errorf("error: vault is already initialized")
		}
	}

	req := &api.InitRequest{}
	if shares, ok := d.GetOk("secret_shares"); ok {
		req.SecretShares = shares.(int)
	}
	if threshold, ok := d.GetOk("secret_threshold"); ok {
		req.SecretThreshold = threshold.(int)
	}
	if keys, ok := d.GetOk("pgp_keys"); ok {
		keylist := []string{}
		kl := keys.([]interface{})
		for _, key := range kl {
			keylist = append(keylist, key.(string))
		}
		req.PGPKeys = keylist
	}

	resp, err := client.Sys().Init(req)
	if err != nil {
		return fmt.Errorf("error: failed to  initialize vault: %s", err)
	}

	d.Set("root_token", resp.RootToken)

	keys := strings.Join(resp.Keys, "\n")
	if err := ioutil.WriteFile(d.Get("unseal_keys_output_file").(string), []byte(keys), 0600); err != nil {
		return fmt.Errorf("failed to write unseal keys: error was: %s. keys are: %s.", err, keys)
	}

	for _, addr := range instances {
		vc.Address = addr

		client, err := api.NewClient(vc)
		if err != nil {
			return err
		}

		sealed, err := client.Sys().SealStatus()
		if err != nil {
			return fmt.Errorf("error: failed to check vault status: %s", err)
		}
		if sealed.Sealed {
			for _, key := range resp.Keys {
				resp, err := client.Sys().Unseal(key)
				if err != nil {
					return err
				}
				if !resp.Sealed {
					break
				}
			}
		}
	}

	d.SetId(uuid.NewV4().String())

	if d.Get("wait_for_ready").(bool) {
		address := d.Get("address").(string)
		if address == "" {
			return errors.New("wait_for_ready requires address to wait for")
		}
		vc.Address = address

		client, err := api.NewClient(vc)
		if err != nil {
			return err
		}
		var lasterr error
		timeout := time.Now().Add(10 * time.Minute)
		for {
			select {
			case <-time.After(10 * time.Second):
				if time.Now().After(timeout) {
					if lasterr == nil {
						return errors.New("error: failed to check vault status")
					}
					return lasterr
				}

				_, err := client.Sys().SealStatus()
				if err != nil {
					lasterr = fmt.Errorf("error: failed to check vault status: %s", err)
				} else {
					return nil
				}
			}
		}
	}
	return nil
}

func resourceInitDelete(d *schema.ResourceData, meta interface{}) error {
	d.SetId("")
	return nil
}

func noop(d *schema.ResourceData, meta interface{}) error {
	return nil
}
