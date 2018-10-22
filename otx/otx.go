/*

*/
package main

import (
	"flag"
	"os"
	"fmt"
	"io/ioutil"
	"time"
	"log"

	"github.com/hosom/gobrointel"
)

const (
	feedName = "OTX"
	desc = "AlienVault OTX Pulse Indicators"
	otxBaseURI = "https://otx.alienvault.com/api/v1/pulses/subscribed"
	otxPulseURI = "https://otx.alienvault.com/pulse/%s"
)

// mapOtxType is a convenience map to map the types used by the OTX API 
// to the types used by the intelligence framework.
var mapOtxType = map[string]brointel.IndicatorType{
	"IPv4": brointel.Addr,
	"IPv6": brointel.Addr,
	"domain": brointel.Domain,
	"hostname": brointel.Domain,
	"email": brointel.Email,
	"URL": brointel.URL,
	"URI": brointel.URL,
	"FileHash-MD5": brointel.FileHash,
	"FileHash-SHA1": brointel.FileHash,
	"FileHash-SHA256": brointel.FileHash,
}

func main() {

	
	apiKey := flag.String("apiKey", "", "API key for accessing OTX.")
	days := flag.Int("days", 30, "How many days of pulses should be downloaded.")
	doNotice := flag.Bool("doNotice", false, 
		"Whether this intel source should generate Notices.")
	flag.Parse()

	// Get x days ago, then convert it to a string for use in the API calls
	today := time.Now()
	searchDate := today.AddDate(0, 0, -*days)
	searchDateStr := searchDate.UTC().Format(time.RFC3339)

	log.Printf("Searching OTX for pulses with starting date of %s", searchDateStr)

	c := NewClient(*apiKey)
	// Date needs to be in ISO8601
	f := c.IterPulses(map[string]string{"limit":"15", "page":"1", "modified_since": searchDateStr})
	
	meta := brointel.MetaData {
		Source: feedName,
		Desc: desc,
		URL: "PlaceHolder",
		DoNotice: *doNotice,
	}

	tmpfile, err := ioutil.TempFile("./", "tempintel")
	if err != nil {
		log.Fatal(err)
	}
	defer tmpfile.Close()
	// print headers to bro intel file
	fmt.Fprintln(tmpfile, brointel.Headers())

	for pul := range f {
		for _, indicator := range pul.Indicators {
			// check if type is supported, some OTX indicators don't 
			// translate to the intel framework.
			if iocType, ok := mapOtxType[*indicator.Type]; ok {
				// update the url to be specific to the pulse
				meta.URL = fmt.Sprintf(otxPulseURI, *pul.ID)

				// build intelligence framework entry
				ioc := brointel.Item{
					Indicator: *indicator.Indicator,
					Type: iocType,
					Meta: meta,
				}

				// print intelligence framework entry
				fmt.Fprintln(tmpfile, ioc.String())	
			}		
		}
	}

	fname := tmpfile.Name()
	err = os.Rename(fname, "./otx.dat")
	if err != nil {
		log.Fatal(err)
	}
}