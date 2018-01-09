package vault

import (
	"encoding/base64"
	"testing"

	"github.com/caiyeon/goldfish/config"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/vault/api"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGoldfishWrapper(t *testing.T) {
	// start vault in dev mode
	cfg, ch, _, wrappingToken, err := config.LoadConfigDev()
	if err != nil {
		panic(err)
	}
	defer close(ch)

	// bootstrap goldfish to vault
	SetConfig(cfg.Vault)
	err = Bootstrap(wrappingToken)
	if err != nil {
		panic(err)
	}

	Convey("Testing bootstrap functions", t, func() {
		Convey("Reusing the server's own token as raw token", func() {
			temp := vaultToken
			unbootstrap()
			err = BootstrapRaw(temp)
			So(err, ShouldBeNil)
		})
		Convey("Bootstrapping via non-approle token", func() {
			rootAuth := &AuthInfo{ID: "goldfish", Type: "token"}

			// create a non-approle wrapped token
			temp := true
			secret, err := rootAuth.CreateToken(
				&api.TokenCreateRequest{
					Policies: []string{"default", "goldfish"},
					Renewable: &temp,
				},
				false, "", "5m",
			)

			So(err, ShouldBeNil)
			So(secret, ShouldNotBeNil)
			So(secret.WrapInfo, ShouldNotBeNil)

			unbootstrap()
			err = Bootstrap(secret.WrapInfo.Token)
			So(err, ShouldBeNil)
		})
	})

	// start unit tests
	Convey("Testing API wrapper", t, func() {
		// this will be imitating the client token
		rootAuth := &AuthInfo{ID: "goldfish", Type: "token"}

		Convey("Server's vault client should not contain a token", func() {
			client, err := NewVaultClient()
			So(err, ShouldBeNil)
			So(client.Token(), ShouldEqual, "")
		})

		// run-time config
		Convey("Config should be loaded", func() {
			c := GetConfig()
			So(c, ShouldResemble, RuntimeConfig{
				ServerTransitKey:  "goldfish",
				UserTransitKey:    "usertransit",
				TransitBackend:    "transit",
				DefaultSecretPath: "secret/",
				BulletinPath:      "secret/bulletins/",
				LastUpdated:       c.LastUpdated,
			})
		})

		// credentials
		Convey("Encrypting and decrypting credentials should work", func() {
			root := rootAuth.ID
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
				So(len(bulletins), ShouldEqual, 4)
				So(bulletins[3], ShouldResemble, map[string]interface{}{
					"title":   "Message title",
					"message": "Message body",
					"type":    "is-success",
				})
			})

			Convey("Listing secrets should work", func() {
				secrets, err := rootAuth.ListSecret("secret/bulletins")
				So(err, ShouldBeNil)
				So(len(secrets), ShouldEqual, 4)
				So(secrets, ShouldContain, "testbulletin")
			})

			Convey("Deleting secrets should work", func() {
				_, err := rootAuth.DeleteSecret("secret/bulletins/testbulletin")
				So(err, ShouldBeNil)

				Convey("Deleted secrets should not be readable anymore", func() {
					resp, err := rootAuth.ReadSecret("secret/bulletins/testbulletin")
					So(err, ShouldNotBeNil)
					So(resp, ShouldBeNil)
				})
			})

			Convey("Wrapping arbitrary data", func() {
				wrapToken, err := rootAuth.WrapData("300s",
					`{ "abc": "def", "ghi": "jkl" }`,
				)
				So(err, ShouldBeNil)
				So(wrapToken, ShouldNotBeBlank)

				// empty auth should still be able to unwrap
				emptyAuth := AuthInfo{}
				resp, err := emptyAuth.UnwrapData(wrapToken)
				So(err, ShouldBeNil)

				data := resp.Data
				So(data, ShouldContainKey, "abc")
				So(data["abc"].(string), ShouldEqual, "def")
				So(data["ghi"].(string), ShouldEqual, "jkl")
			})
		})

		// tokens
		Convey("Creating a token", func() {
			resp, err := rootAuth.CreateToken(&api.TokenCreateRequest{}, false, "", "")
			So(err, ShouldBeNil)
			So(len(resp.Auth.ClientToken), ShouldEqual, 36)

			tempAuth := &AuthInfo{ID: resp.Auth.ClientToken, Type: "token"}

			Convey("Number of accessors should increase", func() {
				accessors, err := rootAuth.GetTokenAccessors()
				So(err, ShouldBeNil)
				So(len(accessors), ShouldEqual, 4)

				_, err = rootAuth.CreateToken(&api.TokenCreateRequest{}, true, "", "")
				So(err, ShouldBeNil)

				accessorsAfter, err := rootAuth.GetTokenAccessors()
				So(err, ShouldBeNil)
				So(len(accessors)+1, ShouldEqual, len(accessorsAfter))
			})

			Convey("With a wrapped ttl", func() {
				resp, err := rootAuth.CreateToken(&api.TokenCreateRequest{}, false, "", "300s")
				So(err, ShouldBeNil)
				So(len(resp.WrapInfo.Token), ShouldEqual, 36)

				// empty auth should still be able to unwrap
				emptyAuth := AuthInfo{}
				resp, err = emptyAuth.UnwrapData(resp.WrapInfo.Token)
				So(err, ShouldBeNil)
				So(len(resp.Auth.ClientToken), ShouldEqual, 36)
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

			Convey("Accessor should be lookup-able", func() {
				resp, err := rootAuth.LookupTokenByAccessor(resp.Auth.Accessor + "," + resp.Auth.Accessor)
				So(err, ShouldBeNil)
				So(len(resp), ShouldEqual, 2)
			})

			Convey("Token should be deleteable via accessor", func() {
				So(rootAuth.RevokeTokenByAccessor(resp.Auth.Accessor), ShouldBeNil)

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
			randomBytes, err := uuid.GenerateRandomBytes(16)
			So(err, ShouldBeNil)
			otp := base64.StdEncoding.EncodeToString(randomBytes)
			status, err := GenerateRootInit(otp)
			So(err, ShouldBeNil)
			So(status.Progress, ShouldEqual, 0)

			// supplying a fake unseal key
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
			// there should be only one user created in PrepareVault()
			_, err = rootAuth.ListUserpassUsers()
			So(err, ShouldBeNil)

			_, err = rootAuth.DeleteRaw("auth/userpass/users/testuser")
			So(err, ShouldBeNil)

			// there should be only one approle (goldfish)
			roles, err := rootAuth.ListApproleRoles()
			So(err, ShouldBeNil)
			So(len(roles), ShouldEqual, 1)

			_, err = rootAuth.DeleteRaw("auth/approle/role/goldfish")
			So(err, ShouldBeNil)
		})

		// roles
		Convey("Listing token roles should work", func() {
			resp, err := rootAuth.ListRoles()
			So(err, ShouldBeNil)
			So(len(resp.([]interface{})), ShouldEqual, 1)

			resp, err = rootAuth.GetRole("testrole")
			So(err, ShouldBeNil)
		})

		// logging in
		Convey("Logging in with different methods", func() {
			resp, err := rootAuth.Login()
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)

			resp, err = (&AuthInfo{ID: "not_a_token", Type: "token"}).Login()
			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)

			resp, err = (&AuthInfo{ID: "fish1", Pass: "golden", Type: "userpass"}).Login()
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)

			resp, err = (&AuthInfo{ID: "fish1", Pass: "notgolden", Type: "userpass"}).Login()
			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)

			resp, err = (&AuthInfo{ID: "tesla", Pass: "password", Type: "ldap"}).Login()
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)

			resp, err = (&AuthInfo{ID: "tesla", Pass: "notpassword", Type: "ldap"}).Login()
			So(err, ShouldNotBeNil)
			So(resp, ShouldBeNil)
		})

		// ldap
		Convey("Listing LDAP groups and users", func() {
			resp, err := rootAuth.ListLDAPGroups()
			So(err, ShouldBeNil)
			So(resp, ShouldResemble, []LDAPGroup{
				LDAPGroup{
					Name:     "engineers",
					Policies: []string{"foobar"},
				},
				LDAPGroup{
					Name:     "scientists",
					Policies: []string{"bar", "foo"},
				},
			})

			resp2, err := rootAuth.ListLDAPUsers()
			So(err, ShouldBeNil)
			So(resp2, ShouldResemble, []LDAPUser{
				LDAPUser{
					Name:     "tesla",
					Policies: []string{"zoobar"},
					Groups:   []string{"engineers"},
				},
			})
		})

	}) // end prepared vault convey

} // end test function
