package config

import (
	"errors"
	"os"

	auditFile "github.com/hashicorp/vault/builtin/audit/file"
	auditSocket "github.com/hashicorp/vault/builtin/audit/socket"
	auditSyslog "github.com/hashicorp/vault/builtin/audit/syslog"

	credAppId "github.com/hashicorp/vault/builtin/credential/app-id"
	credAppRole "github.com/hashicorp/vault/builtin/credential/approle"
	credAws "github.com/hashicorp/vault/builtin/credential/aws"
	credCert "github.com/hashicorp/vault/builtin/credential/cert"
	credGitHub "github.com/hashicorp/vault/builtin/credential/github"
	credLdap "github.com/hashicorp/vault/builtin/credential/ldap"
	credOkta "github.com/hashicorp/vault/builtin/credential/okta"
	credRadius "github.com/hashicorp/vault/builtin/credential/radius"
	credUserpass "github.com/hashicorp/vault/builtin/credential/userpass"

	"github.com/hashicorp/vault/builtin/logical/aws"
	"github.com/hashicorp/vault/builtin/logical/cassandra"
	"github.com/hashicorp/vault/builtin/logical/consul"
	"github.com/hashicorp/vault/builtin/logical/database"
	"github.com/hashicorp/vault/builtin/logical/mongodb"
	"github.com/hashicorp/vault/builtin/logical/mssql"
	"github.com/hashicorp/vault/builtin/logical/mysql"
	"github.com/hashicorp/vault/builtin/logical/pki"
	"github.com/hashicorp/vault/builtin/logical/postgresql"
	"github.com/hashicorp/vault/builtin/logical/rabbitmq"
	"github.com/hashicorp/vault/builtin/logical/ssh"
	"github.com/hashicorp/vault/builtin/logical/totp"
	"github.com/hashicorp/vault/builtin/logical/transit"

	"github.com/hashicorp/vault/audit"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/command"
	"github.com/hashicorp/vault/meta"
	"github.com/hashicorp/vault/api"
	"github.com/mitchellh/cli"
)

func SetupVault(addr, rootToken string) error {
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

	// write sample bulletins
	if _, err := client.Logical().Write("secret/bulletins/bulletina", map[string]interface{}{
		"message": "hello world",
		"title":   "sampleBulletinA",
		"type":    "is-success",
	}); err != nil {
		return err
	}
	if _, err := client.Logical().Write("secret/bulletins/bulletinb", map[string]interface{}{
		"message": "this is sample b",
		"title":   "sampleBulletinB",
		"type":    "is-success",
	}); err != nil {
		return err
	}
	if _, err := client.Logical().Write("secret/bulletins/bulletinc", map[string]interface{}{
		"message": "this is sample c",
		"title":   "sampleBulletinc",
		"type":    "is-success",
	}); err != nil {
		return err
	}

	// todo: write sample users

	// todo: mount pki backend

	return nil
}

func initDevVaultCore() chan struct{} {
	ui := &cli.BasicUi{
		Reader: os.Stdin,
		Writer: os.Stdout,
	}
	m := meta.Meta{
		Ui: ui,
		TokenHelper: command.DefaultTokenHelper,
	}
	shutdownCh := make(chan struct{})

	go (&command.ServerCommand{
		Meta: m,
		AuditBackends: map[string]audit.Factory{
			"file":   auditFile.Factory,
			"syslog": auditSyslog.Factory,
			"socket": auditSocket.Factory,
		},
		CredentialBackends: map[string]logical.Factory{
			"approle":  credAppRole.Factory,
			"cert":     credCert.Factory,
			"aws":      credAws.Factory,
			"app-id":   credAppId.Factory,
			"github":   credGitHub.Factory,
			"userpass": credUserpass.Factory,
			"ldap":     credLdap.Factory,
			"okta":     credOkta.Factory,
			"radius":   credRadius.Factory,
		},
		LogicalBackends: map[string]logical.Factory{
			"aws":        aws.Factory,
			"consul":     consul.Factory,
			"postgresql": postgresql.Factory,
			"cassandra":  cassandra.Factory,
			"pki":        pki.Factory,
			"transit":    transit.Factory,
			"mongodb":    mongodb.Factory,
			"mssql":      mssql.Factory,
			"mysql":      mysql.Factory,
			"ssh":        ssh.Factory,
			"rabbitmq":   rabbitmq.Factory,
			"database":   database.Factory,
			"totp":       totp.Factory,
		},
		ShutdownCh: shutdownCh,
		SighupCh:   command.MakeSighupCh(),
	}).Run([]string{
		"-dev",
		"-dev-listen-address=127.0.0.1:8200",
		"-dev-root-token-id=goldfish",
	})

	return shutdownCh
}

func generateWrappedSecretID(v VaultConfig, token string) (string, error) {
	client, err := api.NewClient(api.DefaultConfig())
	if err := client.SetAddress(v.Address); err != nil {
		return "", err
	}
	client.SetToken(token)
	client.SetWrappingLookupFunc(func(operation, path string) string {
		return "5m"
	})

	resp, err := client.Logical().Write("auth/approle/role/goldfish/secret-id", map[string]interface{}{})
	if err != nil {
		return "", err
	}

	if resp == nil || resp.WrapInfo == nil || resp.WrapInfo.Token == "" {
		return "", errors.New("Failed to setup vault client")
	}

	return resp.WrapInfo.Token, nil
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
