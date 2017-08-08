package request

import (
	"testing"
	"fmt"
	"github.com/caiyeon/goldfish/config"
	"github.com/caiyeon/goldfish/vault"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRequestSystem(t *testing.T) {
	// start vault in dev mode
	cfg, ch, unsealTokens, wrappingToken, err := config.LoadConfigDev()
	if err != nil {
		panic(err)
	}
	defer close(ch)

	// bootstrap goldfish to vault
	vault.SetConfig(cfg.Vault)
	err = vault.StartGoldfishWrapper(wrappingToken)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Unseal tokens: %q", unsealTokens)

	Convey("Testing request system", t, func() {
		// this will be imitating the client token
		rootAuth := &vault.AuthInfo{ID: "goldfish", Type: "token"}

		Convey("Testing policy requests", func() {
			// adding a policy req with a new policy name
			hash, err := Add(rootAuth, map[string]interface{}{
				"Type":       "policy",
				"policyname": "abc",
				"rules":      "# this is a sample policy rule",
			})
			So(err, ShouldBeNil)
			So(hash, ShouldNotBeEmpty)

			// retrieve the request
			req, err := Get(rootAuth, hash)
			So(err, ShouldBeNil)
			polreq, ok := req.(*PolicyRequest)
			So(ok, ShouldEqual, true)
			So(polreq, ShouldNotBeEmpty)
		})
	})
}
