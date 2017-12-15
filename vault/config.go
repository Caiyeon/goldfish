package vault

import (
	"encoding/json"
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/fatih/structs"
	"github.com/mitchellh/hashstructure"
)

type RuntimeConfig struct {
	ServerTransitKey  string
	UserTransitKey    string
	TransitBackend    string
	DefaultSecretPath string
	BulletinPath      string

	SlackWebhook string
	SlackChannel string

	GithubAccessToken  string
	GithubRepoOwner    string
	GithubRepo         string
	GithubPoliciesPath string

	// fields that goldfish will write
	LastUpdated         string `hash:"ignore"`
}

var (
	conf                       = RuntimeConfig{}
	configLock                 = new(sync.RWMutex)
	configHash          uint64 = 0
)

func GetConfig() RuntimeConfig {
	configLock.RLock()
	defer configLock.RUnlock()
	return conf
}

func loadConfigFromVault(path string) error {
	client, err := NewGoldfishVaultClient()
	if err != nil {
		return err
	}

	resp, err := client.Logical().Read(path)
	if err != nil {
		return err
	} else if resp == nil {
		return errors.New("Failed to read config secret from vault")
	}

	// marshall into temp config to ensure it is valid
	temp := RuntimeConfig{}
	if b, err := json.Marshal(resp.Data); err == nil {
		if err := json.Unmarshal(b, &temp); err != nil {
			return err
		}
	} else {
		return err
	}

	// improperly formed slack webhooks are not allowed
	if !strings.HasPrefix(temp.SlackWebhook, "https://hooks.slack.com/services") {
		temp.SlackWebhook = ""
		temp.SlackChannel = ""
	}

	// don't waste a lock if nothing has changed
	newHash, err := hashstructure.Hash(temp, nil)
	if err != nil {
		return err
	}
	if newHash == configHash {
		return nil
	}

	// timestamp the change in vault, notifying operators that the config has been updated
	// if timestamp can't be written, operation should be aborted
	temp.LastUpdated = time.Now().Format(time.UnixDate)
	_, err = client.Logical().Write(path, structs.Map(temp))
	if err != nil {
		return errors.New("Goldfish could not write to runtime config path: " + err.Error())
	}

	// RWLock.Lock() will block read lock requests until it is done
	configLock.Lock()
	defer configLock.Unlock()

	conf = temp
	configHash = newHash

	log.Println("[INFO ]: Server config reloaded")
	return nil
}
