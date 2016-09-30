package main

import (
	"flag"
	"log"
	"runtime"
	"time"
)

var (
	Version string
	Build   string
)

// Stats channel buffer
const statsChBuf = 10

// Configuration structure shared by all files.
type Config struct {
	DownloadDir    string `json:"download_dir"`
	ThreadsPerCore int    `json:"ThreadsPerCore"`
	RunInterval    string `json:"RunInterval"`
	Feeds          []struct {
		URL          string `json:"url"`
		NumDownloads int    `json:"num_downloads"`
	} `json:"feeds"`
}

func systemProfile(tPerCore int) int {
	log.Printf("Entering systemProfile with %v threads per core", tPerCore)
	var numCores int
	var numRoutines int

	numCores = runtime.NumCPU()
	log.Printf("Current system has %d cores", numCores)
	numRoutines = numCores * tPerCore
	log.Printf("Setting concurrency to %d", numRoutines)

	log.Printf("Exiting systemProfile")
	return numRoutines
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting podfetch")
	log.Printf("Version: %s, build: %s", Version, Build)

	// Grab config file location as cmd line arg.
	cfgPtr := flag.String("C", "REQUIRED", "Path to application configuration")
	flag.Parse()
	p := &ConfigLocation{}
	p.Path = *cfgPtr
	log.Printf("Reading configuration from %s", p.Path)

	// Read config from disk
	// If initial load fails, exit since we have nothing to go on.
	err := loadConfig(p)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Get access to valid config pointer
	c := GetConfig()
	err = c.Validate()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	log.Printf("Download_dir %s", c.DownloadDir)
	log.Printf("Threads per core %d", c.ThreadsPerCore)
	log.Printf("Run interval %s", c.RunInterval)

	for _, e := range c.Feeds {
		log.Println(e.URL)
		log.Println(e.NumDownloads)
	}

	duration, err := time.ParseDuration(c.RunInterval)
	if err != nil {
		panic(err)
	}
	log.Printf("Ticker interval set to: %v", duration)

	log.Printf("Setting GOMAXPROCS to %d", c.ThreadsPerCore)
	runtime.GOMAXPROCS(systemProfile(c.ThreadsPerCore))

	dlLogChan := make(chan *DownloadLog, 10)
	go statsInit(dlLogChan)

	log.Printf("Intial fetch")
	fetchManager(dlLogChan)

	log.Printf("Starting timer")
	timer := time.NewTimer(duration)

	// Process events
	for {
		select {
		case <-timer.C:
			log.Println("Timer fired")

			fetchManager(dlLogChan)
			log.Printf("Restarting timer")
			timer = time.NewTimer(duration)
		}
	}

}
