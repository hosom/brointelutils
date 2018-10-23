/*

*/

package main

import (
	"net/http"
	"time"
	"bufio"
	"strings"
	"io/ioutil"
	"log"
	"fmt"
	"os"

	"github.com/hosom/gobrointel"
)

const (
	baseURI = "https://ransomwaretracker.abuse.ch/"
	feedName = "Abuse.ch Ransomware Tracker"
	desc = "Detect ransomware botnet C&C"
	urlBL = "downloads/RW_URLBL.txt"
	domainBL = "downloads/RW_DOMBL.txt"
)

func getIOCs(uri string) <-chan string {

	c := make(chan string)

	go func() {
		defer close(c)
		
		netClient := &http.Client{
			Timeout: time.Second * 10,
		}

		resp, err := netClient.Get(uri)
		if err != nil {
			log.Fatal("Failed to retrieve intelligence feed.")
		}
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			entry := scanner.Text()
			// skip and don't process comments
			if strings.HasPrefix(entry, "#") {
				continue
			}
			// skip and don't process empty lines
			if entry == "" {
				continue
			}
			c <- entry
		}
	}()

	return c
}

func main() {

	// create temporary file
	tmpfile, err := ioutil.TempFile("./", "tempintel")
	if err != nil {
		log.Fatal(err)
	}
	defer tmpfile.Close()

	meta := brointel.MetaData{
		Source: feedName, 
		Desc: desc, 
		URL: baseURI, 
		DoNotice: true,
	}
	
	// print Bro intel headers to the temp file
	fmt.Fprintln(tmpfile, brointel.Headers())

	uri := baseURI + domainBL
	for entry := range getIOCs(uri) {
		ioc := brointel.Item{
			Indicator: entry,
			Type: brointel.Domain,
			Meta: meta,
		}
		fmt.Fprintln(tmpfile, ioc.String())
	}

	uri = baseURI + urlBL
	for entry := range getIOCs(uri) {
		ioc := brointel.Item{
			Indicator: entry,
			Type: brointel.Domain,
			Meta: meta,
		}
		fmt.Fprintln(tmpfile, ioc.String())
	}

	// rename intel file for Bro to consume
	fname := tmpfile.Name()
	err = os.Rename(fname, "./ransomware.dat")
	if err != nil {
		log.Fatal(err)
	}
}