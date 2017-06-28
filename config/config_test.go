package config

import (
	"testing"

	"github.com/hashicorp/vault/api"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConfigParser(t *testing.T) {
	Convey("Parser should accept valid string", t, func() {
		cfg, err := ParseConfig(configString)
		So(cfg, ShouldNotBeNil)
		So(err, ShouldBeNil)
		validateConfig(cfg)
	})

	Convey("Parser should reject invalid strings - no listener config", t, func() {
		cfg, err := ParseConfig(`
			# no listener config
			vault {
				address       = "http://127.0.0.8200"
			}
			`)
		So(cfg, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})

	Convey("Parser should reject invalid strings - no vault config", t, func() {
		cfg, err := ParseConfig(`
			listener "tcp" {
				address       = "127.0.0.1:8000"
			}
			# no vault config
			`)
		So(cfg, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})

	Convey("Starting up a dev vault", t, func() {
		cfg, shutdownCh, secretID, err := LoadConfigDev()
		validateConfig(cfg)
		So(shutdownCh, ShouldNotBeNil)
		defer close(shutdownCh)
		So(secretID, ShouldNotBeNil)
		So(err, ShouldBeNil)
		validateVaultHealth(cfg)
	})

	Convey("Loading custom config", t, func() {
		cfg, err := LoadConfigFile("sample.hcl")
		validateConfig(cfg)
		So(err, ShouldBeNil)
	})
}

func validateConfig(cfg *Config) {
	So(cfg, ShouldNotBeNil)
	cfgListener := cfg.Listener
	So(cfgListener, ShouldNotBeNil)
	So(cfgListener.Type, ShouldEqual, "tcp")
	So(cfgListener.Address, ShouldEqual, "127.0.0.1:8000")
	So(cfgListener.Tls_disable, ShouldBeTrue)
	So(cfgListener.Tls_cert_file, ShouldBeEmpty)
	So(cfgListener.Tls_key_file, ShouldBeEmpty)
	So(cfgListener.Tls_autoredirect, ShouldBeFalse)

	cfgVault := cfg.Vault
	So(cfgVault, ShouldNotBeNil)
	So(cfgVault.Type, ShouldEqual, "vault")
	So(cfgVault.Address, ShouldEqual, "http://127.0.0.1:8200")
	So(cfgVault.Tls_skip_verify, ShouldBeFalse)
	So(cfgVault.Runtime_config, ShouldEqual, "secret/goldfish")
	So(cfgVault.Approle_login, ShouldEqual, "auth/approle/login")
	So(cfgVault.Approle_id, ShouldEqual, "goldfish")
}

func validateVaultHealth(cfg *Config) {
	client, err := api.NewClient(api.DefaultConfig())
	So(client, ShouldNotBeNil)
	So(err, ShouldBeNil)
	client.SetAddress(cfg.Vault.Address)

	sys := client.Sys()
	So(sys, ShouldNotBeNil)
	resp, err := sys.Health()
	So(resp, ShouldNotBeNil)
	So(err, ShouldBeNil)
	So(resp.Initialized, ShouldBeTrue)
	So(resp.Sealed, ShouldBeFalse)
	So(resp.Standby, ShouldBeFalse)
}

const configString = `
listener "tcp" {
	address          = "127.0.0.1:8000"
	tls_cert_file    = ""
	tls_key_file     = ""
	tls_disable      = 1
	tls_autoredirect = 0
}
vault {
	address         = "http://127.0.0.1:8200"
	tls_skip_verify = 0
	runtime_config  = "secret/goldfish"
	approle_login   = "auth/approle/login"
	approle_id      = "goldfish"
}
`
