package vault

import (
	"encoding/json"
	"sync"
	"time"
	"log"
	"reflect"
)

type Config struct {
	ServerTransitKey  string
	UserTransitKey    string
	TransitBackend    string
	DefaultSecretPath string
}

var config Config
var configLock = new(sync.RWMutex)
var LastUpdated time.Time

func GetConfig() Config {
	configLock.RLock()
	defer configLock.RUnlock()
	return config
}

func loadConfigFromVault(path string) error {
	resp, err := vaultClient.Logical().Read(path)
	if err != nil {
		return err
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

	// don't waste a lock if nothing has changed
	if reflect.DeepEqual(temp, config) {
		return nil
	}

	// RWLock.Lock() will block read lock requests until it is done
	configLock.Lock()
	defer configLock.Unlock()

	config = temp
	LastUpdated = time.Now()
	log.Println("Goldfish config reloaded")

	return nil
}