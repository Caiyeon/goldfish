package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/vault/helper/parseutil"
)

type Config struct {
	Listener        *ListenerConfig `hcl:"-"`
	Vault           *VaultConfig    `hcl:"-"`
	DisableMlock    bool            `hcl:"-"`
	DisableMlockRaw interface{}     `hcl:"disable_mlock"`
}

type ListenerConfig struct {
	Type             string
	Address          string
	Tls_disable      bool
	Tls_cert_file    string
	Tls_key_file     string
	Tls_autoredirect bool
}

type VaultConfig struct {
	Type            string
	Address         string
	Tls_skip_verify bool
	Runtime_config  string
	Approle_login   string
	Approle_id      string
}

func LoadConfigFile(path string) (*Config, error) {
	if path == "" {
		return nil, errors.New("[ERROR]: Config file not specified")
	}
	d, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseConfig(string(d))
}

func LoadConfigDev() (*Config, chan struct{}, []string, string, error) {
	// start a vault dev instance
	unsealToken, shutdownCh := initDevVaultCore()

	// multiple unseal tokens would more accurately represent a prod vault system
	unsealTokens, err := rekeyDevVault(unsealToken, 5, 3)
	if err != nil {
		return nil, nil, nil, "", err
	}

	// setup local vault instance with required mounts
	err = setupVault("http://127.0.0.1:8200", "goldfish")
	if err != nil {
		return nil, nil, nil, "", err
	}

	// setup goldfish internal config
	u, _ := url.Parse("http://127.0.0.1:8200")

	result := Config{
		Listener: &ListenerConfig{
			Type:        "tcp",
			Address:     "127.0.0.1:8000",
			Tls_disable: true,
		},
		Vault: &VaultConfig{
			Type:           "vault",
			Address:        u.String(),
			Runtime_config: "secret/goldfish",
			Approle_login:  "auth/approle/login",
			Approle_id:     "goldfish",
		},
		DisableMlock: true,
	}

	// generate an approle secret ID
	secretID, err := generateWrappedSecretID(*result.Vault, "goldfish")
	if err != nil {
		return nil, nil, nil, "", err
	}

	return &result, shutdownCh, unsealTokens, secretID, nil
}

func ParseConfig(d string) (*Config, error) {
	// parse as hcl
	obj, err := hcl.Parse(d)
	if err != nil {
		return nil, err
	}

	// make a new config and decode hcl
	result := Config{
		Listener: &ListenerConfig{},
		Vault:    &VaultConfig{},
	}
	if err := hcl.DecodeObject(&result, obj); err != nil {
		return nil, err
	}

	// config file should have a root object
	list, ok := obj.Node.(*ast.ObjectList)
	if !ok {
		return nil, errors.New("[ERROR]: Config file doesn't have a root object")
	}

	// perform checks on root config keys
	if result.DisableMlockRaw != nil {
		if result.DisableMlock, err = parseutil.ParseBool(result.DisableMlockRaw); err != nil {
			return nil, err
		}
	}

	// config root object should contain only this set of keys
	valid := []string{
		"listener",
		"vault",
		"disable_mlock",
	}
	if err := checkHCLKeys(list, valid); err != nil {
		return nil, err
	}

	// build each specific config component
	if object := list.Filter("listener"); len(object.Items) != 1 {
		return nil, fmt.Errorf("Config requires exactly one 'listener' object")
	} else {
		// there should be only one listener to parse
		if err := parseListener(&result, object.Items[0]); err != nil {
			return nil, fmt.Errorf("Error parsing 'listener': %s", err.Error())
		}
	}

	if object := list.Filter("vault"); len(object.Items) != 1 {
		return nil, fmt.Errorf("Config requires exactly one 'vault' object")
	} else {
		// there should be only one vault to parse
		if err := parseVault(&result, object.Items[0]); err != nil {
			return nil, fmt.Errorf("Error parsing 'vault': %s", err.Error())
		}
	}

	return &result, nil
}

func checkHCLKeys(node ast.Node, valid []string) error {
	var list *ast.ObjectList
	switch n := node.(type) {
	case *ast.ObjectList:
		list = n
	case *ast.ObjectType:
		list = n.List
	default:
		return fmt.Errorf("Cannot recognize HCL key type %T", n)
	}

	validMap := make(map[string]struct{}, len(valid))
	for _, v := range valid {
		validMap[v] = struct{}{}
	}

	var err *multierror.Error
	for _, item := range list.Items {
		key := item.Keys[0].Token.Value().(string)
		if _, ok := validMap[key]; !ok {
			err = multierror.Append(err, fmt.Errorf("Invalid key '%s' on line '%d'", key, item.Assign.Line))
		}
	}
	return err.ErrorOrNil()
}

func parseListener(result *Config, listener *ast.ObjectItem) error {
	key := "listener"
	if len(listener.Keys) > 0 {
		key = listener.Keys[0].Token.Value().(string)
	}

	valid := []string{
		"address",
		"tls_disable",
		"tls_cert_file",
		"tls_key_file",
		"tls_autoredirect",
	}
	if err := checkHCLKeys(listener.Val, valid); err != nil {
		return fmt.Errorf("listener.%s: %s", key, err.Error())
	}

	var m map[string]string
	if err := hcl.DecodeObject(&m, listener.Val); err != nil {
		return fmt.Errorf("listener.%s: %s", key, err.Error())
	}

	// check and enforce field values
	result.Listener.Type = strings.ToLower(key)

	if address, ok := m["address"]; !ok || address == "" {
		return fmt.Errorf("listener.%s: address is required", key)
	} else {
		result.Listener.Address = address
	}

	if certFile, ok := m["tls_cert_file"]; ok {
		result.Listener.Tls_cert_file = certFile
	}
	if keyFile, ok := m["tls_key_file"]; ok {
		result.Listener.Tls_key_file = keyFile
	}

	if tlsDisable, ok := m["tls_disable"]; ok {
		if tlsDisable == "1" {
			result.Listener.Tls_disable = true
		} else if tlsDisable != "0" {
			return fmt.Errorf("listener.%s: tls_disable can be 0 or 1", key)
		}
	}

	if redirect, ok := m["tls_autoredirect"]; ok {
		if redirect == "1" {
			if result.Listener.Tls_disable {
				return fmt.Errorf("listener.%s: tls_autoredirect conflicts with tls_disable", key)
			}
			result.Listener.Tls_autoredirect = true
		} else if redirect != "0" {
			return fmt.Errorf("listener.%s: tls_autoredirect can be 0 or 1", key)
		}
	}

	return nil
}

func parseVault(result *Config, vault *ast.ObjectItem) error {
	key := "vault"
	if len(vault.Keys) > 0 {
		key = vault.Keys[0].Token.Value().(string)
	}

	valid := []string{
		"address",
		"tls_skip_verify",
		"runtime_config",
		"approle_login",
		"approle_id",
	}
	if err := checkHCLKeys(vault.Val, valid); err != nil {
		return fmt.Errorf("vault.%s: %s", key, err.Error())
	}

	var m map[string]string
	if err := hcl.DecodeObject(&m, vault.Val); err != nil {
		return fmt.Errorf("vault.%s: %s", key, err.Error())
	}

	// check and enforce field values, possibly writing default values
	result.Vault.Type = strings.ToLower(key)

	if address, ok := m["address"]; !ok || address == "" {
		return fmt.Errorf("vault.%s: address is required", key)
	} else {
		if url, err := url.Parse(address); err != nil {
			return fmt.Errorf("failed to set address %v reason: %s", address, err.Error())
		} else {
			if !(url.Scheme == "http" || url.Scheme == "https") {
				return fmt.Errorf("vault.%s: address must be prefixed with scheme i.e. http:// or https://", key)
			}
			result.Vault.Address = url.String()
		}
	}

	if tlsSkip, ok := m["tls_skip_verify"]; ok {
		if tlsSkip == "1" {
			result.Vault.Tls_skip_verify = true
		} else if tlsSkip != "0" {
			return fmt.Errorf("vault.%s: tls_disable can be 0 or 1", key)
		}
	}

	if runtimeConfig, ok := m["runtime_config"]; ok && runtimeConfig != "" {
		result.Vault.Runtime_config = runtimeConfig
	} else {
		result.Vault.Runtime_config = "secret/goldfish"
	}

	if login, ok := m["approle_login"]; ok && login != "" {
		result.Vault.Approle_login = login
	} else {
		result.Vault.Approle_login = "auth/approle/login"
	}

	if id, ok := m["approle_id"]; ok && id != "" {
		result.Vault.Approle_id = id
	} else {
		result.Vault.Approle_id = "goldfish"
	}

	return nil
}
