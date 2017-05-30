/*
Convert Abuse.ch tracker data to Bro intel framework 
*/

package main

import (
	"os"
	"fmt"
	"flag"
	"net/http"
	"time"
	"strings"
	"log"
	"bufio"

	"github.com/hosom/gobrointel"
)

const (
	zeusBaseURI = "https://zeustracker.abuse.ch/blocklist.php?download=%s"
)

func usage() {
	fmt.Printf("Usage: %s [OPTIONS] feed_name\n", os.Args[0])
	fmt.Printf(`Possible Feed Values:
domainblocklist 	This blocklist contains the same data as the ZeuS domain 
			blocklist (BadDomains) but with the slight difference that 
			it doesn't exclude hijacked websites (level 2). This means 
			that this blocklist contains all domain names associated 
			with ZeuS C&Cs which are currently being tracked by ZeuS 
			Tracker. Hence this blocklist will likely cause 
			some false positives.

ipblocklist		This blocklist contains the same data as the ZeuS IP blocklist 
			(BadIPs) but with the slight difference that it doesn't exclude 
			hijacked websites (level 2) and free web hosting providers 
			(level 3). This means that this blocklist contains all IPv4 
			addresses associated with ZeuS C&Cswhich are currently being 
			tracked by ZeuS Tracker. Hence this blocklist will likely 
			cause some false positives.

compromised		This blocklist only contains compromised / hijacked websites 
			(level 2) which are being abused by cybercriminals to host a 
			ZeuS botnet controller. Since blocking the FQDN or IP address 
			of compromised host would cause a lot of false positives, the 
			ZeuS compromised URL blocklist contains the full URL to the 
			ZeuS config, dropzone or malware binary instead of the 
			FQDN / IP address.
`)
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	// Remap the help text for the application
	flag.Usage = usage

	flag.Parse()

	args := flag.Args()

	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	// Feed Name is the first positional argument
	feed := args[0]
	// Build URI with the feed name
	uri := fmt.Sprintf(zeusBaseURI, feed)

	var iocType brointel.IndicatorType

	switch feed {
		case "domainblocklist":
			iocType = brointel.Domain
		case "ipblocklist":
			iocType = brointel.Addr
		case "compromised":
			iocType = brointel.URL
		default:
			usage()
	}

	desc := fmt.Sprintf("Zeus Tracker %s feed", feed)
	meta := brointel.MetaData{feed, desc, "https://zeustracker.abuse.ch/blocklist.php", true}

	netClient := &http.Client {
		// Modify go's default 0 timeout
		Timeout: time.Second * 10,
	}

	resp, err := netClient.Get(uri)
	if err != nil {
			log.Fatal("Failed to retrieve intelligence feed.")
	}

	defer resp.Body.Close()

	// Print the intelligence file headers
	fmt.Println(brointel.Headers())

	if resp.StatusCode == 200 {
		// scanner is used to read line delimited data
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			entry := scanner.Text()
			if strings.HasPrefix(entry, "#") {
				continue
			}

			if entry == "" {
				continue
			}

			ioc := brointel.Item{entry, iocType, meta}
			fmt.Println(ioc.String())
		}
	}
}