package request

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/caiyeon/goldfish/vault"
	"github.com/fatih/structs"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/vault/helper/xor"
	"github.com/mitchellh/hashstructure"
	"github.com/mitchellh/mapstructure"
)

// operations on the same request should not interweave,
// a map will prevent this race condition
var lockMap sync.Mutex
var lockHash = make(map[string]bool)

// only one goroutine should perform vault root generation at a time
var lockRoot sync.Mutex

type Request interface {
	IsRootOnly() bool
	Verify(*vault.AuthInfo) error
	Approve(string, string) error
	Reject(*vault.AuthInfo, string) error
}

// adds a request if user has authentication
func Add(auth *vault.AuthInfo, raw map[string]interface{}) (string, error) {
	lockMap.Lock()
	defer lockMap.Unlock()

	t := ""
	if typeRaw, ok := raw["Type"]; !ok {
		if typeRaw, ok = raw["type"]; ok {
			t, _ = typeRaw.(string)
		}
	} else {
		t, _ = typeRaw.(string)
	}
	if t == "" {
		return "", errors.New("Type field is empty")
	}

	switch strings.ToLower(t) {
	case "policy":
		// construct request fields
		req, hash, err := CreatePolicyRequest(auth, raw)
		if err != nil {
			return "", err
		}

		// lock hash in map before writing to vault cubbyhole
		if _, locked := lockHash[hash]; locked {
			return "", errors.New("Someone else is currently editing this request")
		}
		lockHash[hash] = true
		defer delete(lockHash, hash)

		_, err = vault.WriteToCubbyhole("requests/"+hash, structs.Map(req))
		return hash, err

	case "github":
		return "", errors.New("Github requests do not need to be added")

	case "token":
		// construct request fields
		req, hash, err := CreateTokenRequest(auth, raw)
		if err != nil {
			return "", err
		}

		// lock hash in map before writing to vault cubbyhole
		if _, locked := lockHash[hash]; locked {
			return "", errors.New("Someone else is currently editing this request")
		}
		lockHash[hash] = true
		defer delete(lockHash, hash)

		_, err = vault.WriteToCubbyhole("requests/"+hash, structs.Map(req))
		return hash, err

	default:
		return "", errors.New("Unsupported request type")
	}
}

// fetches a request if it exists, and if user has authentication
func Get(auth *vault.AuthInfo, hash string) (Request, error) {
	// lock hash in map before reading from vault cubbyhole
	lockMap.Lock()
	defer lockMap.Unlock()
	if _, locked := lockHash[hash]; locked {
		return nil, errors.New("Someone else is currently editing this request")
	}
	lockHash[hash] = true
	defer delete(lockHash, hash)

	// fetch request from cubbyhole, if it exists
	resp, err := vault.ReadFromCubbyhole("requests/" + hash)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		// a nil response could mean this is a github request
		if len(hash) == 40 {
			if req, err := CreateGithubRequest(auth, map[string]interface{}{
				"commithash": hash,
			}); err != nil {
				return nil, err
			} else {
				_, err = vault.WriteToCubbyhole("requests/"+hash, structs.Map(req))
				return req, nil
			}
		}
		// otherwise, this request simply doesn't exist
		return nil, errors.New("Request ID not found")
	}

	// decode secret to a request
	t := ""
	if typeRaw, ok := resp.Data["Type"]; !ok {
		if typeRaw, ok = resp.Data["type"]; ok {
			t, _ = typeRaw.(string)
		}
	} else {
		t, _ = typeRaw.(string)
	}

	switch strings.ToLower(t) {
	case "policy":
		// decode secret into policy request
		var req PolicyRequest
		if err := mapstructure.Decode(resp.Data, &req); err != nil {
			return nil, err
		}
		// verify hash
		hash_uint64, err := hashstructure.Hash(req, nil)
		if err != nil || strconv.FormatUint(hash_uint64, 16) != hash {
			return nil, errors.New("Hashes do not match")
		}
		// verify policy request is still valid
		if err := req.Verify(auth); err != nil {
			return nil, err
		}
		return &req, nil

	case "github":
		// decode secret into github request
		var req GithubRequest
		if err := mapstructure.Decode(resp.Data, &req); err != nil {
			return nil, err
		}
		// verify user has vault privilege to read the contained policies
		if err := req.Verify(auth); err != nil {
			return nil, err
		}
		return &req, nil

	case "token":
		// decode secret into token creation request
		var req TokenRequest
		if err := mapstructure.Decode(resp.Data, &req); err != nil {
			return nil, err
		}
		// verify user has at least default policy
		if err := req.Verify(auth); err != nil {
			return nil, err
		}
		return &req, nil

	default:
		return nil, errors.New("Invalid request type: " + t)
	}
}

// if unseal is nonempty string, approve request with current auth
// otherwise, add unseal to list of unseals to generate root token later
func Approve(auth *vault.AuthInfo, hash string, unseal string) (Request, error) {
	// lock hash in map before writing to vault cubbyhole
	lockMap.Lock()
	defer lockMap.Unlock()
	if _, locked := lockHash[hash]; locked {
		return nil, errors.New("Someone else is currently editing this request")
	}
	lockHash[hash] = true
	defer delete(lockHash, hash)

	// fetch request from cubbyhole
	resp, err := vault.ReadFromCubbyhole("requests/" + hash)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("Request ID not found")
	}

	// decode secret to a request
	t := ""
	if typeRaw, ok := resp.Data["Type"]; !ok {
		if typeRaw, ok = resp.Data["type"]; ok {
			t, _ = typeRaw.(string)
		}
	} else {
		t, _ = typeRaw.(string)
	}
	if t == "" {
		return nil, errors.New("Invalid request type")
	}

	switch strings.ToLower(t) {
	case "policy":
		// decode secret into policy request
		var req PolicyRequest
		if err := mapstructure.Decode(resp.Data, &req); err != nil {
			return nil, err
		}
		// verify hash
		hash_uint64, err := hashstructure.Hash(req, nil)
		if err != nil || strconv.FormatUint(hash_uint64, 16) != hash {
			return nil, errors.New("Hashes do not match")
		}
		// verify policy request is still valid
		if err := req.Verify(auth); err != nil {
			return nil, err
		}
		if err := req.Approve(hash, unseal); err != nil {
			return nil, err
		}
		return &req, nil

	case "github":
		// decode secret into github request
		var req GithubRequest
		if err := mapstructure.Decode(resp.Data, &req); err != nil {
			return nil, err
		}
		// verify user has vault privileges to read contained policies
		if err := req.Verify(auth); err != nil {
			return nil, err
		}
		if err := req.Approve(hash, unseal); err != nil {
			return nil, err
		}
		return &req, nil

	case "token":
		// decode secret into token creation request
		var req TokenRequest
		if err := mapstructure.Decode(resp.Data, &req); err != nil {
			return nil, err
		}
		// verify user has at least default policy
		if err := req.Verify(auth); err != nil {
			return nil, err
		}
		if err := req.Approve(hash, unseal); err != nil {
			return nil, err
		}
		return &req, nil

	default:
		return nil, errors.New("Invalid request type: " + t)
	}
}

// deletes request, if user is authorized to read resource
func Reject(auth *vault.AuthInfo, hash string) error {
	// lock hash in map before writing to vault cubbyhole
	lockMap.Lock()
	defer lockMap.Unlock()
	if _, locked := lockHash[hash]; locked {
		return errors.New("Someone else is currently editing this request")
	}
	lockHash[hash] = true
	defer delete(lockHash, hash)

	// fetch request from cubbyhole
	resp, err := vault.ReadFromCubbyhole("requests/" + hash)
	if err != nil {
		return err
	}
	if resp == nil {
		return errors.New("Request ID not found")
	}

	// decode secret to a request
	t := ""
	if typeRaw, ok := resp.Data["Type"]; !ok {
		if typeRaw, ok = resp.Data["type"]; ok {
			t, _ = typeRaw.(string)
		}
	} else {
		t, _ = typeRaw.(string)
	}
	if t == "" {
		return errors.New("Invalid request type")
	}

	// verify user can access resource
	switch strings.ToLower(t) {
	case "policy":
		// decode secret into policy request
		var req PolicyRequest
		if err := mapstructure.Decode(resp.Data, &req); err != nil {
			return err
		}
		// verify hash
		hash_uint64, err := hashstructure.Hash(req, nil)
		if err != nil || strconv.FormatUint(hash_uint64, 16) != hash {
			return errors.New("Hashes do not match")
		}
		// verify policy request is still valid
		return req.Reject(auth, hash)

	case "github":
		// decode secret into github request
		var req GithubRequest
		if err := mapstructure.Decode(resp.Data, &req); err != nil {
			return err
		}
		// verify user has vault privileges to read contained policies
		if err := req.Verify(auth); err != nil {
			return err
		}
		return req.Reject(auth, hash)

	case "token":
		// decode secret into token creation request
		var req TokenRequest
		if err := mapstructure.Decode(resp.Data, &req); err != nil {
			return err
		}
		// verify user has at least default policy
		if err := req.Verify(auth); err != nil {
			return err
		}
		return req.Reject(auth, hash)

	default:
		return errors.New("Invalid request type: " + t)
	}
}

func IsRootOnly(req Request) bool {
	return req.IsRootOnly()
}

// attempts to generate a root token via unseal keys
// will return error if another key generation process is underway
func generateRootToken(unsealKeys []string) (string, error) {
	lockRoot.Lock()
	defer lockRoot.Unlock()

	// initialize root generation with a randomly generated otp
	randomBytes, err := uuid.GenerateRandomBytes(16)
	if err != nil {
		return "", err
	}
	otp := base64.StdEncoding.EncodeToString(randomBytes)
	status, err := vault.GenerateRootInit(otp)
	if err != nil {
		return "", err
	}

	if status.EncodedRootToken == "" {
		for _, s := range unsealKeys {
			status, err = vault.GenerateRootUpdate(s, status.Nonce)
			// an error likely means one of the unseals was not valid
			if err != nil {
				errS := "Could not generate root token: " + err.Error()
				// try to cancel the root generation
				if err := vault.GenerateRootCancel(); err != nil {
					errS += ". Attempted to cancel root generation, but: " + err.Error()
				}
				return "", errors.New(errS)
			}
		}
	}

	if status.EncodedRootToken == "" {
		return "", errors.New("Could not generate root token. Was vault re-keyed just now?")
	}

	tokenBytes, err := xor.XORBase64(status.EncodedRootToken, otp)
	if err != nil {
		return "", errors.New("Could not decode root token. Please search and revoke")
	}

	token, err := uuid.FormatUUID(tokenBytes)
	if err != nil {
		return "", errors.New("Could not decode root token. Please search and revoke")
	}

	return token, nil
}

// writes the provided unseal in and returns a slice of all unseals in hash
func appendUnseal(hash, unseal string) ([]string, error) {
	// read current request from cubbyhole
	resp, err := vault.ReadFromCubbyhole("unseal_wrapping_tokens/" + hash)
	if err != nil {
		return nil, err
	}

	var wrappingTokens []string

	// if there are already unseals, read them and append
	if resp != nil {
		raw := ""
		if temp, ok := resp.Data["wrapping_tokens"]; ok {
			raw, _ = temp.(string)
		}
		if raw == "" {
			return nil, errors.New("Could not find key 'wrapping_tokens' in cubbyhole")
		}
		wrappingTokens = append(wrappingTokens, strings.Split(raw, ";")...)
	}

	// wrap the unseal token
	newWrappingToken, err := vault.WrapData("60m", map[string]interface{}{
		"unseal_token": unseal,
	})
	if err != nil {
		return nil, err
	}

	// add the new unseal key in
	wrappingTokens = append(wrappingTokens, newWrappingToken)

	// write the unseals back to the cubbyhole
	_, err = vault.WriteToCubbyhole("unseal_wrapping_tokens/"+hash,
		map[string]interface{}{
			"wrapping_tokens": strings.Trim(strings.Join(strings.Fields(fmt.Sprint(wrappingTokens)), ";"), "[]"),
		},
	)
	return wrappingTokens, err
}

func unwrapUnseals(wrappingTokens []string) (unseals []string, err error) {
	for _, wrappingToken := range wrappingTokens {
		data, err := vault.UnwrapData(wrappingToken)
		if err != nil {
			return nil, err
		}
		if unseal, ok := data["unseal_token"]; ok {
			unseals = append(unseals, unseal.(string))
		} else {
			return nil, errors.New("One of the wrapping tokens timed out. Progress reset.")
		}
	}
	return
}
