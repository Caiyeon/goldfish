// credits mainly go to contributors of
// github.com/hashicorp/vault/command/server/config.go

// this file is purposefully built to follow suit of vault's config parsing

package main

import (
	"log"
	"io/ioutil"
	"fmt"
	"errors"
	"strings"

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
)

var ch = make(chan error)

// Config is the configuration for the goldfish server
type Config struct {
	Listener *Listener `hcl:"-"`
	Vault    *Vault    `hcl:"-"`
}

type Listener struct {
	Type   string
	Config map[string]string
}

type Vault struct {
	Type   string
	Config map[string]string
}



func LoadConfigFile(path string) (*Config, error) {
	d, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseConfig(string(d))
}

func ParseConfig(d string) (*Config, error) {
	// parse as hcl
	obj, err := hcl.Parse(d)
	if err != nil {
		return nil, err
	}

	var result Config
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

	log.Println(result.Listener)
	log.Println(result.Vault)
	return nil, nil
}

func main() {
	go func () {
		for err := range ch {
			if err != nil {
				log.Println("[ERROR]: ", err.Error())
			}
		}
	}()

	log.Println(LoadConfigFile("config.hcl"))
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
		return fmt.Errorf("listeners.%s: %s", key, err.Error())
	}

	var m map[string]string
	if err := hcl.DecodeObject(&m, listener.Val); err != nil {
		return fmt.Errorf("listeners.%s: %s", key, err.Error())
	}

	result.Listener = &Listener{
		Type:   strings.ToLower(key),
		Config: m,
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

	result.Vault = &Vault{
		Type:   strings.ToLower(key),
		Config: m,
	}
	return nil
}
