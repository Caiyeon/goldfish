package vault

import (
	"encoding/base64"
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
	"github.com/hashicorp/vault/api"

	vaultcore "github.com/hashicorp/vault/vault"
	log "github.com/mgutz/logxi/v1"
	"github.com/mitchellh/cli"
	"github.com/gorilla/securecookie"

	. "github.com/smartystreets/goconvey/convey"
)

func WithPreparedVault(t *testing.T, f func(addr, root, wrappingToken string)) func() {
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
			"TransitBackend=transit",
			"UserTransitKey=usertransit",
			"ServerTransitKey=goldfish",
			"DefaultSecretPath=secret/",
			"BulletinPath=secret/bulletins/",
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

		// return address, root token, and goldfish's token in a wrapping token
		f(addr, result.RootToken, token)
	}
}

func TestGoldfishWrapper(t *testing.T) {

Convey("Launching goldfish with vault instance", t, WithPreparedVault(t,
func(addr, root, wrappingToken string) {
	// make sure vault was started properly
	So(len(root), ShouldEqual, 36)
	So(len(wrappingToken), ShouldEqual, 36)
	fmt.Println("Started vault core with root token:", root)

	// setup cmd line args
	VaultSkipTLS = false
	VaultAddress = addr
	ConfigPath   = "secret/goldfish"

	// function will output the token accessor
	err := StartGoldfishWrapper(
		wrappingToken,
		"goldfish",
		"auth/approle/login",
	)
	So(err, ShouldBeNil)

	// test loading config from secret path
	errorChannel := make(chan error)
	err = LoadConfig(true, errorChannel)
	So(err, ShouldBeNil)
	go func() {
		for err := range errorChannel {
			So(err, ShouldBeNil)
		}
	}()

	// this will be imitating the client token
	rootAuth := &AuthInfo{ID: root, Type: "token"}

	// run-time config
	Convey("Config should be loaded", func() {
		c := GetConfig()
		So(c, ShouldResemble, Config{
			ServerTransitKey  : "goldfish",
			UserTransitKey    : "usertransit",
			TransitBackend    : "transit",
			DefaultSecretPath : "secret/",
			BulletinPath      : "secret/bulletins/",
			LastUpdated       : c.LastUpdated,
		})
	})

	// credentials
	Convey("Encrypting and decrypting credentials should work", func() {
		So(rootAuth.EncryptAuth(), ShouldBeNil)
		So(rootAuth.DecryptAuth(), ShouldBeNil)
		So(rootAuth.ID, ShouldEqual, root)
	})

	// secrets
	Convey("Writing secrets should work", func() {
		resp, err := rootAuth.WriteSecret("secret/bulletins/testbulletin",
			"{\"title\": \"Message title\", \"message\": \"Message body\"," +
			"\"type\": \"is-success\"}",
		)
		So(err, ShouldBeNil)
		So(resp, ShouldBeNil)

		Convey("Reading secrets should work", func() {
			resp, err := rootAuth.ReadSecret("secret/bulletins/testbulletin")
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp["title"].(string), ShouldEqual, "Message title")
			So(resp["message"].(string), ShouldEqual, "Message body")
			So(resp["type"].(string), ShouldEqual, "is-success")
		})

		Convey("Reading bulletins should work", func() {
			bulletins, err := rootAuth.GetBulletins()
			So(err, ShouldBeNil)
			So(len(bulletins), ShouldEqual, 1)
			So(bulletins[0], ShouldResemble, map[string]interface{}{
				"title": "Message title",
				"message": "Message body",
				"type": "is-success",
			})
		})

		Convey("Listing secrets should work", func() {
			secrets, err := rootAuth.ListSecret("secret/bulletins")
			So(err, ShouldBeNil)
			So(len(secrets), ShouldEqual, 1)
			So(secrets[0], ShouldEqual, "testbulletin")
		})
	})

	// tokens
	Convey("Creating a token", func() {
		request := &api.TokenCreateRequest{}
		resp, err := rootAuth.CreateToken(request, "")
		So(err, ShouldBeNil)
		So(len(resp.Auth.ClientToken), ShouldEqual, 36)

		Convey("With a wrapped ttl", func() {
			resp, err := rootAuth.CreateToken(request, "300s")
			So(err, ShouldBeNil)
			So(len(resp.WrapInfo.Token), ShouldEqual, 36)

			// SoonTM
			// Convey("And unwrapping that wrapped token", func() {})
		})

		Convey("Token should be able to clear self", func() {
			tempAuth := &AuthInfo{ID: resp.Auth.ClientToken, Type: "token"}
			tempAuth.Clear()
			So(tempAuth, ShouldResemble, &AuthInfo{})
		})

		Convey("Token should be able to revoke self", func() {
			tempAuth := &AuthInfo{ID: resp.Auth.ClientToken, Type: "token"}
			So(tempAuth.RevokeSelf(), ShouldBeNil)
		})

		Convey("Token should be able to lookup self", func() {
			tempAuth := &AuthInfo{ID: resp.Auth.ClientToken, Type: "token"}
			_, err := tempAuth.LookupSelf()
			So(err, ShouldBeNil)
		})

		Convey("Token should be able to renew self", func() {
			tempAuth := &AuthInfo{ID: resp.Auth.ClientToken, Type: "token"}
			_, err := tempAuth.RenewSelf()
			So(err, ShouldNotBeNil)
		})

		Convey("Token should be deleteable via accessor", func() {
			So(rootAuth.DeleteUser("token", resp.Auth.Accessor), ShouldBeNil)
			tempAuth := &AuthInfo{ID: resp.Auth.ClientToken, Type: "token"}
			_, err := tempAuth.LookupSelf()
			So(err, ShouldNotBeNil)
			_, err = tempAuth.RenewSelf()
			So(err, ShouldNotBeNil)
		})
	})

	// mounts
	Convey("Mount operations", func() {
		resp, err := rootAuth.ListMounts()
		So(err, ShouldBeNil)
		So(len(resp), ShouldEqual, 4) // transit, secret, sys, cubbyhole

		settings, err := rootAuth.GetMount("secret")
		So(err, ShouldBeNil)
		So(settings, ShouldNotBeNil)

		// writing a mount's settings again will actually trigger a proper vault write
		So(rootAuth.TuneMount("secret", api.MountConfigInput{
			DefaultLeaseTTL: "",
			MaxLeaseTTL:     "",
		}), ShouldBeNil)
	})

	// helper functions
	Convey("Helper functions should not return errors if vault is healthy", func() {
		// state checks
		_, err = VaultHealth()
		So(err, ShouldBeNil)
		_, err = GenerateRootStatus()
		So(err, ShouldBeNil)

		// generating a new root token
		otp := base64.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(16))
		status, err := GenerateRootInit(otp)
		So(err, ShouldBeNil)
		So(status.Progress, ShouldEqual, 0)

		// supplying a fake unseal token
		status, err = GenerateRootUpdate("YWJjZGVmZ2hpamtsbW5vcHFyc3Q=", status.Nonce)
		So(err, ShouldBeNil)
		So(status.Progress, ShouldEqual, 1)

		// cancelling unseal process
		So(GenerateRootCancel(), ShouldBeNil)

		// cubbyhole operations
		_, err = WriteToCubbyhole("testsecret", map[string]interface{}{
			"key": "value",
		})
		So(err, ShouldBeNil)

		resp, err := ReadFromCubbyhole("testsecret")
		So(err, ShouldBeNil)
		So(resp.Data["key"].(string), ShouldEqual, "value")

		_, err = DeleteFromCubbyhole("testsecret")
		So(err, ShouldBeNil)

		// server operations
		So(renewServerToken(), ShouldBeNil)

		wrap, err := WrapData("300s", map[string]interface{}{
			"key": "value",
		})
		So(err, ShouldBeNil)
		So(len(wrap), ShouldEqual, 36)

		wrappedData, err := UnwrapData(wrap)
		So(err, ShouldBeNil)
		So(wrappedData["key"].(string), ShouldEqual, "value")
	})

	// transit
	Convey("Transit functionality should work", func() {
		cipher, err := rootAuth.EncryptTransit("value")
		So(err, ShouldBeNil)

		plaintext, err := rootAuth.DecryptTransit(cipher)
		So(err, ShouldBeNil)
		So(plaintext, ShouldEqual, "value")
	})

	// policies
	Convey("Policy wrappers should work", func() {
		policies, err := rootAuth.ListPolicies()
		So(err, ShouldBeNil)
		So(policies, ShouldContain, "goldfish")

		details, err := rootAuth.GetPolicy("goldfish")
		So(err, ShouldBeNil)
		So(details, ShouldNotBeBlank)

		So(rootAuth.PutPolicy("testpolicy", "# this is an empty policy"), ShouldBeNil)

		So(rootAuth.DeletePolicy("testpolicy"), ShouldBeNil)
	})

})) // end prepared vault convey

} // end test function
