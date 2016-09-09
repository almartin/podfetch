package main

import (
	"encoding/json"
	"os"
	"sync"
)

var (
	config     *Config
	configLock = new(sync.RWMutex)
)

// config path to be provided by cmd line arg.
type ConfigLocation struct {
	Path string
}

func loadConfig(p *ConfigLocation) error {
	// open input file
	fi, err := os.Open(p.Path)
	if err != nil {
		panic(err)
	}

	// close fi on exit and check for its returned error
	defer func() {
		if err := fi.Close(); err != nil {
			panic(err)
		}
	}()

	tmp := new(Config)

	decoder := json.NewDecoder(fi)
	err = decoder.Decode(&tmp)
	if err != nil {
		return err

	}

	// Grab mutex on config, swap pointers.
	configLock.Lock()
	config = tmp
	configLock.Unlock()

	return nil
}

func GetConfig() *Config {
	configLock.RLock()
	defer configLock.RUnlock()
	return config
}
