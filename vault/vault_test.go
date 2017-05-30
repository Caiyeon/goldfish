package vault

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/vault/builtin/credential/approle"
	"github.com/hashicorp/vault/builtin/logical/transit"
	"github.com/hashicorp/vault/command"
	"github.com/hashicorp/vault/helper/logformat"
	"github.com/hashicorp/vault/http"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/meta"
	"github.com/hashicorp/vault/physical"

	vaultcore "github.com/hashicorp/vault/vault"
	log "github.com/mgutz/logxi/v1"
	"github.com/mitchellh/cli"

	. "github.com/smartystreets/goconvey/convey"
)

func WithPreparedVault(t *testing.T, f func(rootToken string)) func() {
	return func() {
		// setup a vault core
		logger := logformat.NewVaultLogger(log.LevelTrace)
		inm := physical.NewInmem(logger)
		coreConfig := &vaultcore.CoreConfig{
			Physical: inm,
			LogicalBackends: map[string]logical.Factory{
				"transit": transit.Factory,
			},
			CredentialBackends: map[string]logical.Factory{
				"approle": approle.Factory,
			},
			DisableMlock: true,
			Seal:         nil,
		}
		core, err := vaultcore.NewCore(coreConfig)
		So(err, ShouldBeNil)

		// ensure core is uninitialized
		init, err := core.Initialized()
		So(err, ShouldBeNil)
		So(init, ShouldEqual, false)

		// initialize vault core
		result, err := core.Initialize(&vaultcore.InitParams{
			BarrierConfig: &vaultcore.SealConfig{
				SecretShares:    5,
				SecretThreshold: 3,
			},
			RecoveryConfig: nil,
		})
		So(err, ShouldBeNil)

		// unseal vault core
		for i := 0; i < 3; i++ {
			_, _ = core.Unseal(result.SecretShares[i])
		}
		status, _ := core.Sealed()
		So(status, ShouldEqual, false)

		// setup http connection and mock ui
		ln, addr := http.TestServer(t, core)
		defer ln.Close()
		ui := new(cli.MockUi)
		m := meta.Meta{
			ClientToken: result.RootToken,
			Ui:          ui,
		}

		// mount transit backend
		c := &command.MountCommand{Meta: m}
		args := []string{
			"-address", addr,
			"transit",
		}
		code := c.Run(args)
		So(code, ShouldEqual, 0)

		// mount approle backend
		c2 := &command.AuthEnableCommand{Meta: m}
		args = []string{
			"-address", addr,
			"approle",
		}
		code = c2.Run(args)
		So(code, ShouldEqual, 0)

		// write goldfish policy
		c3 := &command.PolicyWriteCommand{Meta: m}
		args = []string{
			"-address", addr,
			"goldfish",
			"../vagrant/policies/goldfish.hcl",
		}
		code = c3.Run(args)
		So(code, ShouldEqual, 0)

		// write goldfish approle
		c4 := &command.WriteCommand{Meta: m}
		args = []string{
			"-address", addr,
			"auth/approle/role/goldfish",
			"role_name=goldfish",
			"secret_id_ttl=5m",
			"token_ttl=480h",
			"token_ttl_max=720h",
			"secret_id_num_uses=1",
			"policies=default,goldfish",
		}
		code = c4.Run(args)
		So(code, ShouldEqual, 0)

		c5 := &command.WriteCommand{Meta: m}
		args = []string{
			"-address", addr,
			"auth/approle/role/goldfish/role-id",
			"role_id=goldfish",
		}
		code = c5.Run(args)
		So(code, ShouldEqual, 0)

		// initialize transit key
		c6 := &command.WriteCommand{Meta: m}
		args = []string{
			"-address", addr,
			"-f",
			"transit/keys/goldfish",
		}
		code = c6.Run(args)
		So(code, ShouldEqual, 0)

		// write goldfish run-time settings
		c7 := &command.WriteCommand{Meta: m}
		args = []string{
			"-address", addr,
			"secret/goldfish",
			"TransitBackend='transit'",
			"UserTransitKey='usertransit'",
			"ServerTransitKey='goldfish'",
			"BulletinPath='secret/bulletins/'",
		}
		code = c7.Run(args)
		So(code, ShouldEqual, 0)

		// fetch a token
		c8 := &command.WriteCommand{Meta: m}
		args = []string{
			"-address", addr,
			"-f",
			"-wrap-ttl=20m",
			"auth/approle/role/goldfish/secret-id",
		}
		code = c8.Run(args)
		So(code, ShouldEqual, 0)
		token := strings.Split(ui.OutputWriter.String(), "wrapping_token:")[1]
		token = strings.TrimSpace(strings.Split(token, "\n")[0])

		// perform convey with root token
		f(token)
	}
}

func TestServer(t *testing.T) {
	Convey("Starting a server", t, WithPreparedVault(t, func(rootToken string) {
		So(len(rootToken), ShouldEqual, 36)
		fmt.Println("Started vault core with root token:", rootToken)
	}))
}
