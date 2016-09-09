package main

import (
	"log"
)

type DownloadLog struct {
	itemType string
	item     string
	data     string
	errors   string
}

func statsInit(cdl <-chan *DownloadLog) {
	log.Printf("Initializing stats routine")
	for {
		select {
		case msg := <-cdl:
			log.Printf("Type: %s Item: %s Status: %s Errors: %s",
				msg.itemType, msg.item, msg.data, msg.errors)
		}
	}
}
