package config

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConfigParser(t *testing.T) {
	Convey("Parser should accept valid string", t, func() {
		cfg, err := ParseConfig(configString)
		So(cfg, ShouldNotBeNil)
		So(err, ShouldBeNil)

		// test cfg's struct data for integrity
	})

	Convey("Parser should reject invalid strings", t, func() {
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
		// test vault server returned by LoadConfigDev()
	})
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
