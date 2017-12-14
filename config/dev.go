package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

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

	physAzure "github.com/hashicorp/vault/physical/azure"
	physConsul "github.com/hashicorp/vault/physical/consul"
	physFile "github.com/hashicorp/vault/physical/file"
	physGCS "github.com/hashicorp/vault/physical/gcs"
	physInmem "github.com/hashicorp/vault/physical/inmem"

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

	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/audit"
	"github.com/hashicorp/vault/command"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/meta"
	"github.com/hashicorp/vault/physical"
	"github.com/mitchellh/cli"
)

func setupVault(addr, rootToken string) error {
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
		"policies":           "default, goldfish",
		"secret_id_num_uses": 1,
		"secret_id_ttl":      "5m",
		"period":             "24h",
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

	// mount userpass and write a test user
	if err := client.Sys().EnableAuthWithOptions("userpass", &api.EnableAuthOptions{
		Type: "userpass",
	}); err != nil {
		return err
	}
	if _, err := client.Logical().Write("auth/userpass/users/fish1", map[string]interface{}{
		"password": "golden",
	}); err != nil {
		return err
	}

	// write a test role
	if _, err := client.Logical().Write("auth/token/roles/testrole", map[string]interface{}{
		"allowed_roles": "abc",
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

	// setup pki backend
	if err := client.Sys().Mount("pki", &api.MountInput{
		Type: "pki",
	}); err != nil {
		return err
	}
	if _, err := client.Logical().Write("pki/root/generate/internal", map[string]interface{}{
		"common_name": "myvault.com",
	}); err != nil {
		return err
	}
	if _, err := client.Logical().Write("pki/config/urls", map[string]interface{}{
		"issuing_certificates":    "http://127.0.0.1:8200/v1/pki/ca",
		"crl_distribution_points": "http://127.0.0.1:8200/v1/pki/crl",
	}); err != nil {
		return err
	}
	if _, err := client.Logical().Write("pki/roles/goldfish", map[string]interface{}{
		"allow_any_name":    "true",
		"allow_subdomains":  "true",
		"allow_baredomains": "true",
		"allow_localhost":   "true",
		"max_ttl":           "3m",
	}); err != nil {
		return err
	}

	// generate a couple of certificates
	if _, err := client.Logical().Write("pki/issue/goldfish", map[string]interface{}{
		"common_name": "localhost",
	}); err != nil {
		return err
	}
	if _, err := client.Logical().Write("pki/issue/goldfish", map[string]interface{}{
		"common_name": "localhost",
	}); err != nil {
		return err
	}

	// mount ldap auth
	if err := client.Sys().EnableAuthWithOptions("ldap", &api.EnableAuthOptions{
		Type: "ldap",
	}); err != nil {
		return err
	}

	// Online LDAP test server
	// http://www.forumsys.com/tutorials/integration-how-to/ldap/online-ldap-test-server/
	// this code is very similar to vault's ldap backend unit test
	if _, err := client.Logical().Write("auth/ldap/config", map[string]interface{}{
		"url":      "ldap://ldap.forumsys.com",
		"userattr": "uid",
		"userdn":   "dc=example,dc=com",
		"groupdn":  "dc=example,dc=com",
		"binddn":   "cn=read-only-admin,dc=example,dc=com",
	}); err != nil {
		return err
	}

	// map some groups to policies (that don't exist)
	if _, err := client.Logical().Write("auth/ldap/groups/scientists", map[string]interface{}{
		"policies": "foo,bar",
	}); err != nil {
		return err
	}
	if _, err := client.Logical().Write("auth/ldap/groups/engineers", map[string]interface{}{
		"policies": "foobar",
	}); err != nil {
		return err
	}

	// map user 'tesla' to a special policy on top of its group policy
	if _, err := client.Logical().Write("auth/ldap/users/tesla", map[string]interface{}{
		"groups":   "engineers",
		"policies": "zoobar",
	}); err != nil {
		return err
	}

	return nil
}

func initDevVaultCore() (string, chan struct{}) {
	// temporarily redirect stdout to capture unseal key
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		panic("error pipe: " + err.Error())
	}
	os.Stdout = w

	ui := &cli.BasicUi{
		Reader: os.Stdin,
		Writer: os.Stdout,
	}
	m := meta.Meta{
		Ui:          ui,
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
		PhysicalBackends: map[string]physical.Factory{
			"azure":                  physAzure.NewAzureBackend,
			"consul":                 physConsul.NewConsulBackend,
			"file":                   physFile.NewFileBackend,
			"file_transactional":     physFile.NewTransactionalFileBackend,
			"gcs":                    physGCS.NewGCSBackend,
			"inmem":                  physInmem.NewInmem,
			"inmem_ha":               physInmem.NewInmemHA,
			"inmem_transactional":    physInmem.NewTransactionalInmem,
			"inmem_transactional_ha": physInmem.NewTransactionalInmemHA,
		},
	}).Run([]string{
		"-dev",
		"-dev-listen-address=127.0.0.1:8200",
		"-dev-root-token-id=goldfish",
	})

	// allow 5 seconds for vault to launch. Check every 200 ms
	ticker := time.NewTicker(time.Millisecond * 50)
	go func() {
		time.Sleep(time.Second * 5)
		ticker.Stop()
	}()

	// if vault is ready before 5 seconds countdown, proceed immediately
	for range ticker.C {
		_, err = http.Get("http://127.0.0.1:8200" + "/v1/sys/health")
		if err == nil {
			break
		}
	}
	if err != nil {
		panic(err)
	}

	// extract unseal key from stdout
	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		var unseal string
		io.Copy(&buf, r)
		leftover := strings.Split(buf.String(), "Unseal Key: ")
		fmt.Sscanf(leftover[1], "%s\n", &unseal)
		outC <- unseal
	}()

	w.Close()
	os.Stdout = old

	return <-outC, shutdownCh
}

func generateWrappedSecretID(v VaultConfig, token string) (string, error) {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return "", err
	}
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

func rekeyDevVault(unsealToken string, shares int, threshold int) ([]string, error) {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return nil, err
	}
	if err := client.SetAddress("http://127.0.0.1:8200"); err != nil {
		return nil, err
	}

	resp, err := client.Sys().RekeyInit(&api.RekeyInitRequest{
		SecretShares:    shares,
		SecretThreshold: threshold,
	})
	if err != nil {
		return nil, err
	}

	resp2, err := client.Sys().RekeyUpdate(unsealToken, resp.Nonce)
	if err != nil {
		return nil, err
	}
	return resp2.Keys, nil
}

const goldfishPolicyRules = `# [mandatory]
# store goldfish run-time settings here
# goldfish hot-reloads from this endpoint every minute
path "secret/goldfish*" {
  capabilities = ["read", "update"]
}


# [optional]
# to enable transit encryption, see wiki for details
path "transit/encrypt/goldfish" {
  capabilities = ["read", "update"]
}
path "transit/decrypt/goldfish" {
  capabilities = ["read", "update"]
}


# [optional]
# for goldfish to fetch certificates from PKI backend
path "pki/issue/goldfish" {
  capabilities = ["update"]
}
`
