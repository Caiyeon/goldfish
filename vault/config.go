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

type Config struct {
	ServerTransitKey    string
	UserTransitKey      string
	TransitBackend      string
	DefaultSecretPath   string
	BulletinPath        string

	SlackWebhook        string
	SlackChannel        string

	GithubAccessToken   string
	GithubRepoOwner     string
	GithubRepo          string
	GithubPoliciesPath  string
	GithubTargetBranch  string

	// fields that goldfish will write
	LastUpdated         string `hash:"ignore"`
	GithubCurrentCommit string
}

var (
	config              = Config{}
	configLock          = new(sync.RWMutex)
	configHash uint64   = 0
	GithubCurrentCommit = ""
)

func GetConfig() Config {
	configLock.RLock()
	defer configLock.RUnlock()
	return config
}

func loadConfigFromVault(path string) error {
	resp, err := vaultClient.Logical().Read(path)
	if err != nil {
		return err
	} else if resp == nil {
		return errors.New("Failed to read config secret from vault")
	}

	// marshall into temp config to ensure it is valid
	temp := Config{}
	if b, err := json.Marshal(resp.Data); err == nil {
		if err := json.Unmarshal(b, &temp); err != nil {
			return err
		}
	} else {
		return err
	}

	// the local copy of current commit is the source of truth
	temp.GithubCurrentCommit = GithubCurrentCommit

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
	_, err = vaultClient.Logical().Write(path, structs.Map(temp))
	if err != nil {
		return errors.New("As of v0.2.3, goldfish needs write permissions to the config_path vault endpoint.")
	}

	// RWLock.Lock() will block read lock requests until it is done
	configLock.Lock()
	defer configLock.Unlock()

	config             = temp
	configHash         = newHash

	log.Println("Goldfish config reloaded")
	return nil
}
