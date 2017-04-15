package vault

import (
	"encoding/json"
	"sync"
	"time"
)

type Config struct {
	ServerTransitKey string
	UserTransitKey   string
	LastUpdated      time.Time
}

var config *Config
var configLock = new(sync.RWMutex)

func GetConfig() *Config {
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
	temp := new(Config)
	if b, err := json.Marshal(resp.Data); err == nil {
		if err := json.Unmarshal(b, &temp); err != nil {
			return err
		}
	} else {
		return err
	}

	// RWLock.Lock() will block read lock requests until it is done
	configLock.Lock()
	defer configLock.Unlock()

	config = temp
	config.LastUpdated = time.Now()

	return nil
}