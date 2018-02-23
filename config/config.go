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
	Type                 string
	Address              string
	Tls_disable          bool
	Tls_autoredirect     bool
	Cert                 *Certificate
	Pki_cert             *Pki_certificate
	Lets_encrypt_address string
}

type Certificate struct {
	Cert_file string
	Key_file  string
}

type Pki_certificate struct {
	Pki_path    string
	Common_name string
	Alt_names   []string
	Ip_sans     []string
}

type VaultConfig struct {
	Type            string
	Address         string
	Tls_skip_verify bool
	Runtime_config  string
	Approle_login   string
	Approle_id      string
	CA_cert			string
	CA_path			string
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
			// Pki_cert:    &Pki_certificate{
			// 	Pki_path: "pki/issue/goldfish",
			// 	Common_name: "localhost",
			// 	Alt_names: []string{"vault-ui.io", "abc.com"},
			// 	Ip_sans: []string{"10.0.0.1", "172.0.0.1", "127.0.0.1"},
			// },
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
	result.Listener.Type = key

	valid := []string{
		"address",
		"tls_disable",
		"tls_autoredirect",
		"certificate",
		"pki_certificate",
		"lets_encrypt",
	}
	if err := checkHCLKeys(listener.Val, valid); err != nil {
		return fmt.Errorf("listener.%s: %s", key, err.Error())
	}

	var temp map[string]interface{}
	if err := hcl.DecodeObject(&temp, listener.Val); err != nil {
		return fmt.Errorf("listener.%s: %s", key, err.Error())
	}

	// map configurations to struct
	if raw, ok := temp["address"]; ok {
		if result.Listener.Address, ok = raw.(string); !ok {
			return fmt.Errorf("listener.%s: address must be a string", key)
		}
	} else {
		return fmt.Errorf("listener.%s: address is required", key)
	}

	if raw, ok := temp["tls_disable"]; ok {
		if b, ok := raw.(int); ok {
			if b == 1 {
				result.Listener.Tls_disable = true
			} else if b != 0 {
				return fmt.Errorf("listener.%s: tls_disable must be 0 or 1", key)
			}
		} else {
			return fmt.Errorf("listener.%s: tls_disable must be 0 or 1", key)
		}
	}

	if raw, ok := temp["tls_autoredirect"]; ok {
		if b, ok := raw.(int); ok {
			if b == 1 {
				result.Listener.Tls_autoredirect = true
			} else if b != 0 {
				return fmt.Errorf("listener.%s: tls_autoredirect must be 0 or 1", key)
			}
		} else {
			return fmt.Errorf("listener.%s: tls_autoredirect must be 0 or 1", key)
		}
	}

	// check for configuration conflicts
	if result.Listener.Tls_disable && result.Listener.Tls_autoredirect {
		return fmt.Errorf("listener.%s: tls_autoredirect conflicts with tls_disable", key)
	}

	cert_options := []string{"certificate", "pki_certificate", "lets_encrypt"}

	if result.Listener.Tls_disable {
		for _, opt := range cert_options {
			if _, exists := temp[opt]; exists {
				return fmt.Errorf("listener.%s: tls_disable conflicts with %s", key, opt)
			}
		}
	}

	if !result.Listener.Tls_disable {
		cert_count := 0
		for _, opt := range cert_options {
			if _, exists := temp[opt]; exists {
				cert_count = cert_count + 1
			}
		}
		if cert_count > 1 {
			return fmt.Errorf("listener.%s: multiple certificates are not supported", key)
		}
		if cert_count < 1 {
			return fmt.Errorf("listener.%s: tls is enabled but no certificate option is provided", key)
		}

		// parse the provided certificate option
		obj, ok := listener.Val.(*ast.ObjectType)
		if !ok {
			return fmt.Errorf("listener.%s: no child objects found, but expected certificates", key)
		}

		list := obj.List
		if object := list.Filter("certificate"); len(object.Items) > 1 {
			return fmt.Errorf("listener.%s: multiple certificates are not supported", key)
		} else if len(object.Items) == 1 {
			result.Listener.Cert = &Certificate{}
			if err := parseCertificate(result.Listener.Cert, object.Items[0]); err != nil {
				return fmt.Errorf("listener.%s: %s", key, err.Error())
			}
		}

		if object := list.Filter("pki_certificate"); len(object.Items) > 1 {
			return fmt.Errorf("listener.%s: multiple certificates are not supported", key)
		} else if len(object.Items) == 1 {
			result.Listener.Pki_cert = &Pki_certificate{}
			if err := parsePkiCertificate(result.Listener.Pki_cert, object.Items[0]); err != nil {
				return fmt.Errorf("listener.%s: %s", key, err.Error())
			}
		}

		if object := list.Filter("lets_encrypt"); len(object.Items) > 1 {
			return fmt.Errorf("listener.%s: multiple certificates are not supported", key)
		} else if len(object.Items) == 1 {
			if err := parseLetsEncrypt(&result.Listener.Lets_encrypt_address, object.Items[0]); err != nil {
				return fmt.Errorf("listener.%s: %s", key, err.Error())
			}
		}
	}

	return nil
}

func parseCertificate(result *Certificate, certificate *ast.ObjectItem) error {
	if result == nil || certificate == nil {
		return fmt.Errorf("Parsing certificate... arguments are nil")
	}

	key := "certificate"
	if len(certificate.Keys) > 0 {
		key = certificate.Keys[0].Token.Value().(string)
	}

	valid := []string{
		"cert_file",
		"key_file",
	}
	if err := checkHCLKeys(certificate.Val, valid); err != nil {
		return fmt.Errorf("certificate.%s: %s", key, err.Error())
	}

	if err := hcl.DecodeObject(result, certificate.Val); err != nil {
		return fmt.Errorf("certificate.%s: %s", key, err.Error())
	}

	if result.Cert_file == "" {
		return fmt.Errorf("certificate.%s: cert_file must be provided", key)
	}
	if result.Key_file == "" {
		return fmt.Errorf("certificate.%s: key_file must be provided", key)
	}

	return nil
}

func parsePkiCertificate(result *Pki_certificate, certificate *ast.ObjectItem) error {
	if result == nil || certificate == nil {
		return fmt.Errorf("Parsing pki_certificate... arguments are nil")
	}

	key := "pki"
	if len(certificate.Keys) > 0 {
		key = certificate.Keys[0].Token.Value().(string)
	}

	valid := []string{
		"pki_path",
		"common_name",
		"alt_names",
		"ip_sans",
	}
	if err := checkHCLKeys(certificate.Val, valid); err != nil {
		return fmt.Errorf("certificate.%s: %s", key, err.Error())
	}

	if err := hcl.DecodeObject(result, certificate.Val); err != nil {
		return fmt.Errorf("certificate.%s: %s", key, err.Error())
	}

	if !strings.Contains(result.Pki_path, "issue") {
		return fmt.Errorf("certificate.%s: pki_path must be a full pki issuing path", key)
	}

	if result.Common_name == "" {
		return fmt.Errorf("certificate.%s: config requires common_name field", key)
	}

	return nil
}

func parseLetsEncrypt(result *string, certificate *ast.ObjectItem) error {
	if result == nil || certificate == nil {
		return fmt.Errorf("Parsing lets_encrypt... arguments are nil")
	}

	key := "lets_encrypt"
	if len(certificate.Keys) > 0 {
		key = certificate.Keys[0].Token.Value().(string)
	}

	valid := []string{
		"address",
	}
	if err := checkHCLKeys(certificate.Val, valid); err != nil {
		return fmt.Errorf("lets_encrypt.%s: %s", key, err.Error())
	}

	var temp map[string]string
	if err := hcl.DecodeObject(&temp, certificate.Val); err != nil {
		return fmt.Errorf("lets_encrypt.%s: %s", key, err.Error())
	}

	if temp["address"] == "" {
		return fmt.Errorf("lets_encrypt.%s: address is mandatory", key)
	}
	*result = temp["address"]

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
		"ca_cert",
		"ca_path",
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
			return fmt.Errorf("vault.%s: tls_skip_verify can be 0 or 1", key)
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

	if cacert, ok := m["ca_cert"]; ok && cacert != "" {
		result.Vault.CA_cert = cacert
	} else {
		result.Vault.CA_cert = ""
	}

	if capath, ok := m["ca_path"]; ok && capath != "" {
		result.Vault.CA_path = capath
	} else {
		result.Vault.CA_path = ""
	}

	return nil
}
