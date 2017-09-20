package config

import (
	"testing"

	"github.com/hashicorp/vault/api"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConfigParser(t *testing.T) {
	Convey("Parser should accept valid string", t, func() {
		cfg, err := ParseConfig(defaultConfigString)
		So(err, ShouldBeNil)
		So(cfg, ShouldNotBeNil)
		So(cfg, ShouldResemble, defaultParsedConfig)
	})

	Convey("Parser should accept valid string - default values", t, func() {
		cfg, err := ParseConfig(`
			listener "tcp" {
				address          = "127.0.0.1:8000"
			}
			vault {
				address         = "http://127.0.0.1:8200"
			}
			`)
		So(err, ShouldBeNil)
		So(cfg, ShouldResemble, &Config {
			Listener: &ListenerConfig {
				Type:             "tcp",
				Address:          "127.0.0.1:8000",
				Tls_disable:      false,
				Tls_cert_file:    "",
				Tls_key_file:     "",
				Tls_autoredirect: false,
			},
			Vault: &VaultConfig {
				Type:            "vault",
				Address:         "http://127.0.0.1:8200",
				Tls_skip_verify: false,
				Runtime_config:  "secret/goldfish",
				Approle_login:   "auth/approle/login",
				Approle_id:      "goldfish",
			},
		})
	})

	Convey("Parser should accept valid string - tls_autoredirect enabled (listener)", t, func() {
		cfg, err := ParseConfig(`
			listener "tcp" {
				address          = "127.0.0.1:8000"
				tls_autoredirect = 1
			}
			vault {
				address         = "http://127.0.0.1:8200"
			}
			`)
		So(err, ShouldBeNil)
		So(cfg, ShouldResemble, &Config {
			Listener: &ListenerConfig {
				Type:        "tcp",
				Address:     "127.0.0.1:8000",
				Tls_disable: false,
				Tls_autoredirect: true,
			},
			Vault: &VaultConfig {
				Type:           "vault",
				Address:        "http://127.0.0.1:8200",
				Runtime_config: "secret/goldfish",
				Approle_login:  "auth/approle/login",
				Approle_id:     "goldfish",
			},
		})
	})

	Convey("Parser should reject invalid keys", t, func() {
		cfg, err := ParseConfig(`
			listener "tcp" {
				address          = "127.0.0.1:8000"
				invalid          = "value"
			}
			vault {
				address         = "http://127.0.0.1:8200"
			}
			`)
		So(cfg, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "Invalid key")
		So(err.Error(), ShouldContainSubstring, "invalid")
	})

	Convey("Parser should accept valid string - tls_skip_verify enabled (listener)", t, func() {
		cfg, err := ParseConfig(`
			listener "tcp" {
				address          = "127.0.0.1:8000"
			}
			vault {
				address         = "http://127.0.0.1:8200"
				tls_skip_verify = 1
			}
			`)
		So(err, ShouldBeNil)
		So(cfg, ShouldResemble, &Config {
			Listener: &ListenerConfig {
				Type:        "tcp",
				Address:     "127.0.0.1:8000",
			},
			Vault: &VaultConfig {
				Type:            "vault",
				Address:         "http://127.0.0.1:8200",
				Tls_skip_verify: true,
				Runtime_config:  "secret/goldfish",
				Approle_login:   "auth/approle/login",
				Approle_id:      "goldfish",
			},
		})
	})

	Convey("Parser should reject invalid strings - no listener config", t, func() {
		cfg, err := ParseConfig(`
			# no listener config
			vault {
				address       = "http://127.0.0.8200"
			}
			`)
		So(err, ShouldNotBeNil)
		So(cfg, ShouldBeNil)
	})

	Convey("Parser should reject invalid strings - multiple listener configs", t, func() {
		cfg, err := ParseConfig(`
			listener "tcp" {
				address          = "127.0.0.1:8000"
			}
			listener "tcp" {
				address          = "127.0.0.1:8001"
			}
			vault {
				address         = "http://127.0.0.1:8200"
			}
			`)
		So(err, ShouldNotBeNil)
		So(cfg, ShouldBeNil)
	})

	Convey("Parser should reject invalid listener - no address", t, func() {
		cfg, err := ParseConfig(`
			listener "tcp" {
			}
			vault {
				address         = "http://127.0.0.1:8200"
			}
			`)
		So(err, ShouldNotBeNil)
		So(cfg, ShouldBeNil)
	})

	Convey("Parser should reject invalid listener - empty (invalid) address", t, func() {
		cfg, err := ParseConfig(`
			listener "tcp" {
				address          = ""
			}
			vault {
				address         = "http://127.0.0.1:8200"
			}
			`)
		So(err, ShouldNotBeNil)
		So(cfg, ShouldBeNil)
	})

	Convey("Parser should reject invalid listener - invalid tls_disable", t, func() {
		cfg, err := ParseConfig(`
			listener "tcp" {
				address          = "127.0.0.1:8000"
				tls_disable      = "invalid"
			}
			vault {
				address         = "http://127.0.0.1:8200"
			}
			`)
		So(err, ShouldNotBeNil)
		So(cfg, ShouldBeNil)
	})

	Convey("Parser should reject invalid listener - invalid tls_autoredirect configuration", t, func() {
		cfg, err := ParseConfig(`
			listener "tcp" {
				address          = "127.0.0.1:8000"
				tls_disable      = 1
				tls_autoredirect = 1
			}
			vault {
				address         = "http://127.0.0.1:8200"
			}
			`)
		So(err, ShouldNotBeNil)
		So(cfg, ShouldBeNil)
	})

	Convey("Parser should reject invalid listener - invalid tls_autoredirect", t, func() {
		cfg, err := ParseConfig(`
			listener "tcp" {
				address          = "127.0.0.1:8000"
				tls_autoredirect = "invalid"
			}
			vault {
				address         = "http://127.0.0.1:8200"
			}
			`)
		So(err, ShouldNotBeNil)
		So(cfg, ShouldBeNil)
	})

	Convey("Parser should reject invalid strings - no vault config", t, func() {
		cfg, err := ParseConfig(`
			listener "tcp" {
				address       = "127.0.0.1:8000"
			}
			# no vault config
			`)
		So(err, ShouldNotBeNil)
		So(cfg, ShouldBeNil)
	})

	Convey("Parser should reject invalid strings - multiple vault configs", t, func() {
		cfg, err := ParseConfig(`
			listener "tcp" {
				address          = "127.0.0.1:8000"
			}
			vault {
				address         = "http://127.0.0.1:8200"
			}
			vault {
				address         = "http://127.0.0.1:8200"
			}
			`)
		So(err, ShouldNotBeNil)
		So(cfg, ShouldBeNil)
	})

	Convey("Parser should reject invalid vault - no address", t, func() {
		cfg, err := ParseConfig(`
			listener "tcp" {
				address          = "127.0.0.1:8000"
			}
			vault {
			}
			`)
		So(err, ShouldNotBeNil)
		So(cfg, ShouldBeNil)
	})

	Convey("Parser should reject invalid vault - empty (invalid) address", t, func() {
		cfg, err := ParseConfig(`
			listener "tcp" {
				address          = "127.0.0.1:8000"
			}
			vault {
				address          = ""
			}
			`)
		So(err, ShouldNotBeNil)
		So(cfg, ShouldBeNil)
	})

	Convey("Parser should reject invalid vault - malformed address", t, func() {
		cfg, err := ParseConfig(`
			listener "tcp" {
				address          = "127.0.0.1:8000"
			}
			vault {
				address          = "cache_object:foo/bar>"
			}
			`)
		So(err, ShouldNotBeNil)
		So(cfg, ShouldBeNil)
	})

	Convey("Parser should reject invalid vault - invalid address", t, func() {
		cfg, err := ParseConfig(`
			listener "tcp" {
				address          = "127.0.0.1:8000"
			}
			vault {
				address          = "invalid"
			}
			`)
		So(err, ShouldNotBeNil)
		So(cfg, ShouldBeNil)
	})

	Convey("Parser should reject invalid vault - invalid tls_skip_verify", t, func() {
		cfg, err := ParseConfig(`
			listener "tcp" {
				address          = "127.0.0.1:8000"
			}
			vault {
				address          = "http://127.0.0.1:8200"
				tls_skip_verify  = "invalid"
			}
			`)
		So(err, ShouldNotBeNil)
		So(cfg, ShouldBeNil)
	})

	Convey("Starting up a dev vault", t, func() {
		cfg, shutdownCh, _, secretID, err := LoadConfigDev()
		So(err, ShouldBeNil)
		So(shutdownCh, ShouldNotBeNil)
		defer close(shutdownCh)
		So(cfg, ShouldResemble, devParsedConfig)
		So(secretID, ShouldNotBeNil)

		// validate health of vault
		client, err := api.NewClient(api.DefaultConfig())
		So(err, ShouldBeNil)
		So(client, ShouldNotBeNil)
		client.SetAddress(cfg.Vault.Address)

		sys := client.Sys()
		resp, err := sys.Health()
		So(err, ShouldBeNil)
		So(resp, ShouldNotBeNil)
		So(resp.Initialized, ShouldBeTrue)
		So(resp.Sealed, ShouldBeFalse)
		So(resp.Standby, ShouldBeFalse)
	})

	Convey("Loading valid custom config", t, func() {
		cfg, err := LoadConfigFile("sample.hcl")
		So(err, ShouldBeNil)
		So(cfg, ShouldResemble, sampleParsedConfig)
	})

	Convey("Loading invalid custom config - no file specified", t, func() {
		cfg, err := LoadConfigFile("")
		So(err, ShouldNotBeNil)
		So(cfg, ShouldBeNil)
	})

	Convey("Loading invalid custom config - non-existant file specified", t, func() {
		cfg, err := LoadConfigFile("fake_sample.hcl")
		So(err, ShouldNotBeNil)
		So(cfg, ShouldBeNil)
	})
}

const defaultConfigString = `
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

var defaultParsedConfig = &Config {
	Listener: &ListenerConfig {
		Type:        "tcp",
		Address:     "127.0.0.1:8000",
		Tls_disable: true,
	},
	Vault: &VaultConfig {
		Type:           "vault",
		Address:        "http://127.0.0.1:8200",
		Runtime_config: "secret/goldfish",
		Approle_login:  "auth/approle/login",
		Approle_id:     "goldfish",
	},
	DisableMlock: false,
}

var devParsedConfig = &Config {
	Listener: &ListenerConfig {
		Type:        "tcp",
		Address:     "127.0.0.1:8000",
		Tls_disable: true,
	},
	Vault: &VaultConfig {
		Type:           "vault",
		Address:        "http://127.0.0.1:8200",
		Runtime_config: "secret/goldfish",
		Approle_login:  "auth/approle/login",
		Approle_id:     "goldfish",
	},
	DisableMlock: true,
}

var sampleParsedConfig = &Config {
	Listener: &ListenerConfig {
		Type:        "tcp",
		Address:     "127.0.0.1:8000",
		Tls_disable: true,
	},
	Vault: &VaultConfig {
		Type:           "vault",
		Address:        "http://127.0.0.1:8200",
		Runtime_config: "secret/goldfish",
		Approle_login:  "auth/approle/login",
		Approle_id:     "goldfish",
	},
	DisableMlock: false,
	DisableMlockRaw: 0,
}
