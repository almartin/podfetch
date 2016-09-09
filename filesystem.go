package main

import (
	"io/ioutil"
	"log"
	"os"
)

// Read in the download directory so we can reap files exceeding the configured max.
// Iterate through directory listing in reverse order to delete oldest files.
func reapFiles(path string, max int) {

	var count = 0

	l, _ := ioutil.ReadDir(path)

	for i := len(l) - 1; i >= 0; i-- {
		//log.Printf("Checking for removable files at %d, and name %s", count, l[i].Name())
		if count > (max - 1) {
			file := path + "/" + l[i].Name()
			log.Printf("Deleting file %s", file)
			os.Remove(file)
		}
		count++
	}
}

// Check to see if file exists.
// Returns true if it does, false if it does not.
func fileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}
