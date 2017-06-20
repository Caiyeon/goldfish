package config

import (
	"io/ioutil"
	"fmt"
	"errors"
	"strings"
	"net/url"

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
)

var ch = make(chan error)

type Config struct {
	Listener *ListenerConfig `hcl:"-"`
	Vault    *VaultConfig    `hcl:"-"`
}

type ListenerConfig struct {
	Type          string
	Address       string
	Tls_disable   bool
	Tls_cert_file string
	Tls_key_file  string
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

func LoadConfigDev() (*Config, chan struct{}, string, error) {
	// start a vault dev instance
	shutdownCh := initDevVaultCore()

	// setup local vault instance with required mounts
	err := SetupVault("http://127.0.0.1:8200", "goldfish")
	if err != nil {
		return nil, nil, "", err
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
	}

	// generate an approle secret ID
	secretID, err := generateWrappedSecretID(*result.Vault, "goldfish")
	if err != nil {
		return nil, nil, "", err
	}

	return &result, shutdownCh, secretID, nil
}

func ParseConfig(d string) (*Config, error) {
	// parse as hcl
	obj, err := hcl.Parse(d)
	if err != nil {
		return nil, err
	}

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

	// config root object should contain only this set of keys
	valid := []string{
		"listener",
		"vault",
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
			return nil, fmt.Errorf("Error parsing 'listener': %s", err)
		}
	}

	if object := list.Filter("vault"); len(object.Items) != 1 {
		return nil, fmt.Errorf("Config requires exactly one 'vault' object")
	} else {
		// there should be only one vault to parse
		if err := parseVault(&result, object.Items[0]); err != nil {
			return nil, fmt.Errorf("Error parsing 'vault': %s", err)
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

	var err error
	for _, item := range list.Items {
		key := item.Keys[0].Token.Value().(string)
		if _, ok := validMap[key]; !ok {
			err = fmt.Errorf("Invalid key '%s' on line '%d'", key, item.Assign.Line)
			ch <- err
		}
	}
	return err
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
			return fmt.Errorf("failed to set address: %v", err)
		} else {
			result.Vault.Address = url.String()
		}
	}

	if tlsSkip, ok := m["tls_skip_verify"]; ok {
		if tlsSkip == "1" {
			result.Vault.Tls_skip_verify = true
		} else if tlsSkip != "0" {
			return fmt.Errorf("listener.%s: tls_disable can be 0 or 1", key)
		}
	}

	if runtimeConfig, ok := m["runtime_config"]; ok {
		result.Vault.Runtime_config = runtimeConfig
	} else {
		result.Vault.Runtime_config = "secret/goldfish"
	}

	if login, ok := m["approle_login"]; ok {
		result.Vault.Approle_login = login
	} else {
		result.Vault.Approle_login = "auth/approle/login"
	}

	if id, ok := m["approle_id"]; ok {
		result.Vault.Approle_id = id
	} else {
		result.Vault.Approle_id = "goldfish"
	}

	return nil
}
