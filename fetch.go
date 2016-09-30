package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

type Feed struct {
	Title        string
	URLs         []string
	NumDownloads int
	MaxDownloads int
}

func fetchManager(dlLogCh chan<- *DownloadLog) {
	log.Println("Entering fetchManager")

	var wg sync.WaitGroup

	c := GetConfig()

	for _, e := range c.Feeds {
		wg.Add(1)
		e := e
		go func() {
			log.Printf("Goroutine initialized for feed %s", e.URL)

			defer wg.Done()

			feed, err := getRSS2(e.URL, dlLogCh)
			if err != nil {
				log.Printf("Unable to fetch %s, %s", e.URL, err)
			} else {
				// Grab the episode count from the config
				feed.NumDownloads = e.NumDownloads

				// Call the feed handler
				handleFeed(feed, c.DownloadDir, dlLogCh)
			}
			log.Printf("Ending goroutine for feed %s", e.URL)
		}()
	}
	wg.Wait()
}

func getRSS2(URL string, dlLogCh chan<- *DownloadLog) (*Feed, error) {
	type Enclosure struct {
		URL string `xml:"url,attr"`
	}

	type Item struct {
		Enclosures Enclosure `xml:"enclosure"`
	}

	type Rss2 struct {
		XMLName  xml.Name `xml:"rss"`
		Version  string   `xml:"version,attr"`
		Title    string   `xml:"channel>title"`
		ItemList []Item   `xml:"channel>item"`
	}

	log.Print("Fetching RSS URL: " + URL)

	d := new(DownloadLog)
	d.itemType = "Feed"
	d.item = URL

	var items = new(Rss2)
	var f = new(Feed)

	response, err := http.Get(URL)
	log.Printf("Recieved response of size: %d", response.ContentLength)
	if err != nil {

		d.errors = fmt.Sprintf("Error downloading feed: %s", err)
		dlLogCh <- d

		return nil, err
	}

	decoded := xml.NewDecoder(response.Body)
	err = decoded.Decode(items)
	if err != nil {

		d.errors = fmt.Sprintf("Error parsing feed XML: %s", err)
		dlLogCh <- d

		return nil, err
	}

	f.Title = items.Title

	for _, e := range items.ItemList {
		f.URLs = append(f.URLs, e.Enclosures.URL)
	}

	d.data = "Successfully downloaded feed"
	dlLogCh <- d

	return f, nil
}

func handleFeed(feed *Feed, dlPath string, dlLogCh chan<- *DownloadLog) {
	log.Printf("Handling the %s feed", feed.Title)

	// Create directories for each title
	path := dlPath + "/" + feed.Title

	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	for i, e := range feed.URLs {

		d := new(DownloadLog)
		d.itemType = "Episode"
		d.item = e

		if i <= (feed.NumDownloads - 1) {

			// Extract the filename from the URL, and create a full path from it
			fileName := getLastTokenAfterSlash(e)
			fileName = path + "/" + fileName

			if !fileExists(fileName) {
				log.Println("File does not exist, initiating download")
				err := downloadFile(e, fileName)
				if err != nil {
					log.Println(err)
					d.errors = fmt.Sprintf("Error downloading file: %s", err)
				} else {
					d.data = fmt.Sprintf("Successfully downloaded file")
				}
			} else {
				d.data = "File exists, not downloading"
			}

		} else {
			break
		}

		// Push accumulated statistics
		dlLogCh <- d

		// Clean up old files per config.
		// Execute asynchronously, don't care about result.
		go reapFiles(path, feed.NumDownloads)
	}
}

// Get the last token after a slash
// Used to extract a file name from a URL
// Ex: http://www.example.com/assests/|file|
func getLastTokenAfterSlash(str string) string {
	tokens := strings.Split(str, "/")
	return tokens[len(tokens)-1]
}

// Download a file from URL to path
func downloadFile(url string, p string) (err error) {
	log.Printf("In downloadFile with %s, and %s", p, url)

	// Create the output file
	out, err := os.Create(p)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
