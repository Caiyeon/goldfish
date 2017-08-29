package request

import (
	"crypto/sha256"
	"fmt"
	"testing"

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
	err = vault.Bootstrap(wrappingToken)
	if err != nil {
		panic(err)
	}

	// this will be imitating the client token
	rootAuth := &vault.AuthInfo{ID: "goldfish", Type: "token"}

	// collect requester's information
	self, err := rootAuth.LookupSelf()
	if err != nil {
		panic(err)
	}
	if self == nil {
		panic("Could not confirm requester identity")
	}
	rootAuthHash := fmt.Sprintf("%x", sha256.Sum256([]byte(self.Data["display_name"].(string))))

	Convey("Testing request system", t, func() {
		Convey("Testing policy requests", func() {
			//-----------------------------------------------------------------
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

			// verify request body
			So(req, ShouldResemble, &PolicyRequest{
				Type:          "policy",
				PolicyName:    "abc",
				Previous:      "",
				Proposed:      "# this is a sample policy rule",
				Requester:     "token",
				RequesterHash: rootAuthHash,
				Required:      3,
				Progress:      0,
			})

			// approve the request
			_, err = Approve(rootAuth, hash, unsealTokens[0])
			So(err, ShouldBeNil)
			_, err = Approve(rootAuth, hash, unsealTokens[1])
			So(err, ShouldBeNil)
			_, err = Approve(rootAuth, hash, unsealTokens[2])
			So(err, ShouldBeNil)

			// confirm changes were made
			rules, err := rootAuth.GetPolicy("abc")
			So(err, ShouldBeNil)
			So(rules, ShouldEqual, "# this is a sample policy rule")

			//-----------------------------------------------------------------
			// request a change to the same (now existing) policy
			hash, err = Add(rootAuth, map[string]interface{}{
				"Type":       "policy",
				"policyname": "abc",
				"rules":      "# this is not the same rule",
			})
			So(err, ShouldBeNil)
			So(hash, ShouldNotBeEmpty)

			// retrieve the request
			req, err = Get(rootAuth, hash)
			So(err, ShouldBeNil)

			// approve the request
			_, err = Approve(rootAuth, hash, unsealTokens[0])
			So(err, ShouldBeNil)
			_, err = Approve(rootAuth, hash, unsealTokens[1])
			So(err, ShouldBeNil)
			_, err = Approve(rootAuth, hash, unsealTokens[2])
			So(err, ShouldBeNil)

			// confirm changes were made
			rules, err = rootAuth.GetPolicy("abc")
			So(err, ShouldBeNil)
			So(rules, ShouldEqual, "# this is not the same rule")

			//-----------------------------------------------------------------
			// approve a request with an invalid unseal token
			hash, err = Add(rootAuth, map[string]interface{}{
				"Type":       "policy",
				"policyname": "abc",
				"rules":      "# this is a new rule",
			})
			So(err, ShouldBeNil)
			So(hash, ShouldNotBeEmpty)

			// retrieve the request
			req, err = Get(rootAuth, hash)
			So(err, ShouldBeNil)

			// approve the request
			_, err = Approve(rootAuth, hash, "NotUnseal")
			So(err, ShouldBeNil)
			_, err = Approve(rootAuth, hash, "NotUnseal")
			So(err, ShouldBeNil)
			_, err = Approve(rootAuth, hash, "NotUnseal")
			So(err, ShouldNotBeNil)

			// confirm changes were NOT made
			rules, err = rootAuth.GetPolicy("abc")
			So(err, ShouldBeNil)
			So(rules, ShouldEqual, "# this is not the same rule")

			//-----------------------------------------------------------------
			// rejecting a request halfway
			hash, err = Add(rootAuth, map[string]interface{}{
				"Type":       "policy",
				"policyname": "abc",
				"rules":      "# this is a new rule",
			})
			So(err, ShouldBeNil)
			So(hash, ShouldNotBeEmpty)

			// retrieve the request
			_, err = Get(rootAuth, hash)
			So(err, ShouldBeNil)

			// approve the request
			_, err = Approve(rootAuth, hash, "NotUnseal")
			So(err, ShouldBeNil)
			_, err = Approve(rootAuth, hash, "NotUnseal")
			So(err, ShouldBeNil)

			// reject the request
			err = Reject(rootAuth, hash)
			So(err, ShouldBeNil)

			// confirm request no longer exists
			req, err = Get(rootAuth, hash)
			So(err, ShouldNotBeNil)
			So(req, ShouldBeNil)

			// confirm changes were NOT made
			rules, err = rootAuth.GetPolicy("abc")
			So(err, ShouldBeNil)
			So(rules, ShouldEqual, "# this is not the same rule")

			//-----------------------------------------------------------------
			// removing an exist policy through request
			hash, err = Add(rootAuth, map[string]interface{}{
				"Type":       "policy",
				"policyname": "abc",
				"rules":      "", // empty rules will mark it for deletion
			})
			So(err, ShouldBeNil)
			So(hash, ShouldNotBeEmpty)

			// retrieve the request
			_, err = Get(rootAuth, hash)
			So(err, ShouldBeNil)

			// approve the request
			_, err = Approve(rootAuth, hash, unsealTokens[0])
			So(err, ShouldBeNil)
			_, err = Approve(rootAuth, hash, unsealTokens[1])
			So(err, ShouldBeNil)
			_, err = Approve(rootAuth, hash, unsealTokens[2])
			So(err, ShouldBeNil)

			// confirm request no longer exists
			req, err = Get(rootAuth, hash)
			So(err, ShouldNotBeNil)
			So(req, ShouldBeNil)

			// confirm policy no longer exists
			rules, err = rootAuth.GetPolicy("abc")
			So(err, ShouldBeNil)
			So(rules, ShouldEqual, "")
			policies, err := rootAuth.ListPolicies()
			So(err, ShouldBeNil)
			So(policies, ShouldNotContain, "abc")
		})
	})
}
