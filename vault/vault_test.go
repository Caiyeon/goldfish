package vault

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/vault/builtin/credential/approle"
	"github.com/hashicorp/vault/builtin/credential/userpass"
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
				"transit":  transit.Factory,
			},
			CredentialBackends: map[string]logical.Factory{
				"approle":  approle.Factory,
				"userpass": userpass.Factory,
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
		fmt.Println(addr)
		var code int

		// REQUIRED -----------------------------------------------
		// mount transit backend
		code = (&command.MountCommand{Meta: m}).Run([]string{
			"-address", addr,
			"transit",
		})
		So(code, ShouldEqual, 0)

		// initialize transit key
		code = (&command.WriteCommand{Meta: m}).Run([]string{
			"-address", addr,
			"-f",
			"transit/keys/goldfish",
		})
		So(code, ShouldEqual, 0)

		// write goldfish policy
		code = (&command.PolicyWriteCommand{Meta: m}).Run([]string{
			"-address", addr,
			"goldfish",
			"../vagrant/policies/goldfish.hcl",
		})
		So(code, ShouldEqual, 0)

		// mount approle auth backend
		code = (&command.AuthEnableCommand{Meta: m}).Run([]string{
			"-address", addr,
			"approle",
		})
		So(code, ShouldEqual, 0)

		// write goldfish approle
		code = (&command.WriteCommand{Meta: m}).Run([]string{
			"-address", addr,
			"auth/approle/role/goldfish",
			"role_name=goldfish",
			"secret_id_ttl=5m",
			"token_ttl=480h",
			"token_ttl_max=720h",
			"secret_id_num_uses=1",
			"policies=default,goldfish",
		})
		So(code, ShouldEqual, 0)
		code = (&command.WriteCommand{Meta: m}).Run([]string{
			"-address", addr,
			"auth/approle/role/goldfish/role-id",
			"role_id=goldfish",
		})
		So(code, ShouldEqual, 0)

		// write goldfish run-time settings
		code = (&command.WriteCommand{Meta: m}).Run([]string{
			"-address", addr,
			"secret/goldfish",
			"TransitBackend=transit",
			"UserTransitKey=usertransit",
			"ServerTransitKey=goldfish",
			"DefaultSecretPath=secret/",
			"BulletinPath=secret/bulletins/",
		})
		So(code, ShouldEqual, 0)

		// fetch a token from approle
		code = (&command.WriteCommand{Meta: m}).Run([]string{
			"-address", addr,
			"-f",
			"-wrap-ttl=20m",
			"auth/approle/role/goldfish/secret-id",
		})
		So(code, ShouldEqual, 0)
		token := strings.Split(ui.OutputWriter.String(), "wrapping_token:")[1]
		token = strings.TrimSpace(strings.Split(token, "\n")[0])

		// OPTIONAL -----------------------------------------------
		// mount userpass auth backend
		code = (&command.AuthEnableCommand{Meta: m}).Run([]string{
			"-address", addr,
			"userpass",
		})
		So(code, ShouldEqual, 0)

		// write a test user
		code = (&command.WriteCommand{Meta: m}).Run([]string{
			"-address", addr,
			"auth/userpass/users/testuser",
			"password=foo",
			"policies=admins",
			"ttl=480h",
			"max_ttl=720h",
		})
		So(code, ShouldEqual, 0)

		// write a test role
		code = (&command.WriteCommand{Meta: m}).Run([]string{
			"-address", addr,
			"auth/token/roles/testrole",
			"allowed_roles=abc",
		})
		So(code, ShouldEqual, 0)

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
	VaultAddress = addr
	VaultSkipTLS = false

	// function will output the token accessor
	err := StartGoldfishWrapper(
		wrappingToken,
		"auth/approle/login",
		"goldfish",
	)
	So(err, ShouldBeNil)

	// test loading config from secret path
	errorChannel := make(chan error)
	err = LoadRuntimeConfig("secret/goldfish")
	So(err, ShouldBeNil)
	go func() {
		for err := range errorChannel {
			So(err, ShouldBeNil)
		}
	}()

	// this will be imitating the client token
	rootAuth := &AuthInfo{ID: root, Type: "token"}

	Convey("Server's vault client should not contain a token", func() {
		client, err := NewVaultClient()
		So(err, ShouldBeNil)
		So(client.Token(), ShouldEqual, "")
	})

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

		Convey("Wrapping arbitrary data", func() {
			wrapToken, err := rootAuth.WrapData("300s",
				"{ \"abc\": \"def\", \"ghi\": \"jkl\" }",
			)
			So(err, ShouldBeNil)
			So(wrapToken, ShouldNotBeBlank)

			data, err := rootAuth.UnwrapData(wrapToken)
			So(err, ShouldBeNil)
			So(data, ShouldContainKey, "abc")
			So(data["abc"].(string), ShouldEqual, "def")
			So(data["ghi"].(string), ShouldEqual, "jkl")
		})
	})

	// tokens
	Convey("Creating a token", func() {
		resp, err := rootAuth.CreateToken(&api.TokenCreateRequest{}, "")
		So(err, ShouldBeNil)
		So(len(resp.Auth.ClientToken), ShouldEqual, 36)

		tempAuth := &AuthInfo{ID: resp.Auth.ClientToken, Type: "token"}

		Convey("List of tokens should increase by one", func() {
			countBefore, err := rootAuth.GetTokenCount()
			So(err, ShouldBeNil)

			_, err = rootAuth.CreateToken(&api.TokenCreateRequest{}, "")
			So(err, ShouldBeNil)

			countAfter, err := rootAuth.GetTokenCount()
			So(err, ShouldBeNil)
			So(countBefore + 1, ShouldEqual, countAfter)
		})

		Convey("With a wrapped ttl", func() {
			resp, err := rootAuth.CreateToken(&api.TokenCreateRequest{}, "300s")
			So(err, ShouldBeNil)
			So(len(resp.WrapInfo.Token), ShouldEqual, 36)

			// SoonTM
			// Convey("And unwrapping that wrapped token", func() {})
		})

		Convey("Token lookup self, renew self, and revoke self", func() {
			_, err := tempAuth.LookupSelf()
			So(err, ShouldBeNil)

			_, err = tempAuth.RenewSelf()
			So(err, ShouldNotBeNil)

			So(tempAuth.RevokeSelf(), ShouldBeNil)
		})

		Convey("Token clear self", func() {
			tempAuth.Clear()
			So(tempAuth, ShouldResemble, &AuthInfo{})
		})

		Convey("Token should be deleteable via accessor", func() {
			So(rootAuth.DeleteUser("token", resp.Auth.Accessor), ShouldBeNil)
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
		So(resp, ShouldContainKey, "transit/")

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
		cipher, err := rootAuth.EncryptTransit("usertransit", "value")
		So(err, ShouldBeNil)

		plaintext, err := rootAuth.DecryptTransit("usertransit", cipher)
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

		details, err = rootAuth.GetPolicy("testpolicy")
		So(err, ShouldBeNil)
		So(details, ShouldEqual, "")
	})

	// users
	Convey("Listing users of all types should work", func() {
		// there should be only two tokens: root and goldfish
		resp, err := rootAuth.ListUsers("token", 0)
		So(err, ShouldBeNil)
		So(len(resp.([]interface{})), ShouldEqual, 2)

		// should be an out of bounds error
		resp, err = rootAuth.ListUsers("token", 300)
		So(err, ShouldNotBeNil)
		So(resp, ShouldBeNil)

		// there should be only one user created in PrepareVault()
		_, err = rootAuth.ListUsers("userpass", 0)
		So(err, ShouldBeNil)
		// So(len(resp.([]interface{})), ShouldEqual, 1)
		err = rootAuth.DeleteUser("userpass", "testuser")
		So(err, ShouldBeNil)

		// there should be only one approle (goldfish)
		_, err = rootAuth.ListUsers("approle", 0)
		So(err, ShouldBeNil)
		// So(len(resp.([]interface{})), ShouldEqual, 1)
		err = rootAuth.DeleteUser("approle", "goldfish")
		So(err, ShouldBeNil)
	})

	// roles
	Convey("Listing token roles should work", func() {
		resp, err := rootAuth.ListRoles()
		So(err, ShouldBeNil)
		So(len(resp.([]interface{})), ShouldEqual, 1)

		resp, err = rootAuth.GetRole("testrole")
		So(err, ShouldBeNil)
		fmt.Println(resp)
	})

	// logging in
	Convey("Logging in with different methods", func() {
		resp, err := rootAuth.Login()
		So(err, ShouldBeNil)
		So(resp, ShouldNotBeNil)

		resp, err = (&AuthInfo{ID: "not_a_token", Type: "token"}).Login()
		So(err, ShouldNotBeNil)
		So(resp, ShouldBeNil)

		resp, err = (&AuthInfo{ID: "testuser", Pass: "foo", Type: "userpass"}).Login()
		So(err, ShouldBeNil)
		So(resp, ShouldNotBeNil)

		resp, err = (&AuthInfo{ID: "testuser", Pass: "foobar", Type: "userpass"}).Login()
		So(err, ShouldNotBeNil)
		So(resp, ShouldBeNil)
	})

})) // end prepared vault convey

} // end test function
