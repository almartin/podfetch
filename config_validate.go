package main

import (
	"errors"
	"log"
	"net/url"
	"time"
)

// Application specific configuration validator.
func (c *Config) Validate() error {

	// Check for required configuration values
	if len(c.DownloadDir) == 0 {
		log.Printf("Required attribute DownloadDir not configured")
		err := errors.New("Required attribute DownloadDir not configured")
		return err
	}

	// Check for optional values correctness
	_, err := time.ParseDuration(c.RunInterval)
	if err != nil {
		log.Printf("Unable to convert value %s to duration", c.RunInterval)
		return err
	}

	// Check for Feed validity.
	for _, e := range c.Feeds {
		_, err := url.Parse(e.URL)
		if err != nil {
			return err
		}
		if e.NumDownloads == 0 {
			err := errors.New("Required argument NumDownloads not present")
			return err
		}
	}

	// No errors captured, return nil.
	return nil
}

func (c *Config) Initialize() error {
	return nil
}
