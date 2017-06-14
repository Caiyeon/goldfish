package main

import (
	"encoding/base64"
	"errors"
	// "fmt"
	"net"
	"log"

	auditFile "github.com/hashicorp/vault/builtin/audit/file"
	auditSocket "github.com/hashicorp/vault/builtin/audit/socket"
	auditSyslog "github.com/hashicorp/vault/builtin/audit/syslog"

	credAppRole "github.com/hashicorp/vault/builtin/credential/approle"
	credUserpass "github.com/hashicorp/vault/builtin/credential/userpass"
	credLdap "github.com/hashicorp/vault/builtin/credential/ldap"
	credGitHub "github.com/hashicorp/vault/builtin/credential/github"

	"github.com/hashicorp/vault/builtin/logical/database"
	"github.com/hashicorp/vault/builtin/logical/pki"
	"github.com/hashicorp/vault/builtin/logical/transit"

	"github.com/hashicorp/vault/http"
	"github.com/hashicorp/vault/audit"
	// "github.com/hashicorp/vault/command"
	"github.com/hashicorp/vault/logical"
	// "github.com/hashicorp/vault/meta"

	"github.com/hashicorp/vault/physical"

	"github.com/hashicorp/vault/helper/logformat"
	vaultcore "github.com/hashicorp/vault/vault"
	api "github.com/hashicorp/vault/api"
	logv1 "github.com/mgutz/logxi/v1"
)

func SetupVaultDev(addr, rootToken string) error {
	// initialize vault with required setup details
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return err
	}
	if err := client.SetAddress(addr); err != nil {
		return err
	}
	client.SetToken(rootToken)

	// setup transit backend
	if err := client.Sys().Mount("transit", &api.MountInput{
		Type: "transit",
	}); err != nil {
		return err
	}

	if _, err := client.Logical().Write(
		"transit/keys/goldfish",
		map[string]interface{}{},
	); err != nil {
		return err
	}

	// write goldfish policy
	if err := client.Sys().PutPolicy("goldfish", goldfishPolicyRules); err != nil {
		return err
	}

	// mount approle and write goldfish approle
	if err := client.Sys().EnableAuthWithOptions("approle", &api.EnableAuthOptions{
		Type: "approle",
	}); err != nil {
		return err
	}

	if _, err := client.Logical().Write("auth/approle/role/goldfish", map[string]interface{}{
		"role_name":          "goldfish",
		"secret_id_ttl":      "5m",
		"token_ttl":          "480h",
		"secret_id_num_uses": 1,
		"policies":           "default, goldfish",
	}); err != nil {
		return err
	}

	if _, err := client.Logical().Write("auth/approle/role/goldfish/role-id", map[string]interface{}{
		"role_id": "goldfish",
	}); err != nil {
		return err
	}

	// write runtime config
	if _, err := client.Logical().Write("secret/goldfish", map[string]interface{}{
		"TransitBackend":    "transit",
		"UserTransitKey":    "usertransit",
		"ServerTransitKey":  "goldfish",
		"DefaultSecretPath": "secret/",
		"BulletinPath":      "secret/bulletins/",
	}); err != nil {
		return err
	}

	// mount userpass
	if err := client.Sys().EnableAuthWithOptions("userpass", &api.EnableAuthOptions{
		Type: "userpass",
	}); err != nil {
		return err
	}

	// create 'goldfish' root token
	if _, err := client.Auth().Token().Create(&api.TokenCreateRequest{
		ID:       "goldfish",
		Policies: []string{"root"},
	}); err != nil {
		return err
	}

	return nil
}

func main() {
	go func () {
		for err := range ch {
			if err != nil {
				log.Println("[ERROR]: ", err.Error())
			}
		}
	}()
	result, listener, err := LoadConfigDev()
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	log.Println(result)
}

func initLocalVault() (net.Listener, string, string, []string, error) {
	// core config
	logger := logformat.NewVaultLogger(logv1.LevelTrace)
	inm := physical.NewInmem(logger)
	coreConfig := &vaultcore.CoreConfig{
		Physical: inm,
		AuditBackends: map[string]audit.Factory{
			"file":   auditFile.Factory,
			"syslog": auditSyslog.Factory,
			"socket": auditSocket.Factory,
		},
		CredentialBackends: map[string]logical.Factory{
			"approle":  credAppRole.Factory,
			"github":   credGitHub.Factory,
			"userpass": credUserpass.Factory,
			"ldap":     credLdap.Factory,
		},
		LogicalBackends: map[string]logical.Factory{
			"pki":        pki.Factory,
			"transit":    transit.Factory,
			"database":   database.Factory,
		},
		DisableMlock: true,
		Seal:         nil,
	}

	// start core
	core, err := vaultcore.NewCore(coreConfig)
	if err != nil {
		return nil, "", "", []string{}, err
	}

	// initialize core
	result, err := core.Initialize(&vaultcore.InitParams{
		BarrierConfig: &vaultcore.SealConfig{
			SecretShares:    5,
			SecretThreshold: 3,
		},
		RecoveryConfig: nil,
	})

	// decode byte arrays into strings
	var unsealTokens = make([]string, 5)
	for i := 0; i < 5; i++ {
		unsealTokens[i] = base64.StdEncoding.EncodeToString(result.SecretShares[i])
	}

	// unseal vault core
	for i := 0; i < 3; i++ {
		if _, err = core.Unseal(result.SecretShares[i]); err != nil {
			return nil, "", "", []string{}, err
		}
	}
	if status, err := core.Sealed(); err != nil {
		return nil, "", "", []string{}, err
	} else if status == true {
		return nil, "", "", []string{}, errors.New("Failed to unseal dev vault core")
	}

	// setup http listener for core
	ln, addr := http.TestServer(nil, core)
	return ln, addr, result.RootToken, unsealTokens, nil
}

const goldfishPolicyRules = `
# [mandatory]
# credential transit key (stores logon tokens)
# NO OTHER POLICY should be able to write to this key
path "transit/encrypt/goldfish" {
  capabilities = ["read", "update"]
}
path "transit/decrypt/goldfish" {
  capabilities = ["read", "update"]
}

# [mandatory] [changable]
# store goldfish run-time settings here
# goldfish hot-reloads from this endpoint every minute
path "secret/goldfish*" {
  capabilities = ["read", "update"]
}
`
