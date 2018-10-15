/*

*/
package main

import (
	"flag"
	"os"
	"fmt"

	"github.com/hosom/gobrointel"
	"github.com/AlienVault-OTX/OTX-Go-SDK/src/otxapi"
)

const (
	feed = "OTX"
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
	doNotice := flag.Bool("doNotice", false, 
		"Whether this intel source should generate Notices.")
	flag.Parse()


	meta := brointel.MetaData{feed, desc, "PlaceHolder", *doNotice}

	os.Setenv("X_OTX_API_KEY", *apiKey)

	client := otxapi.NewClient(nil)
	opt := &otxapi.ListOptions{Page: 1, PerPage: 50}
	pulseList, _, err := client.ThreatIntel.List(opt)

	if err != nil {
		fmt.Printf("Error: %v\n\n", err)
	} 

	fmt.Println(brointel.Headers())
	// bool used to signal whether or not the API has more pages of results
	moreResults := true
	for moreResults {
		if pulseList.NextPageString == nil {
			moreResults = false
		}

		for _, pulse := range pulseList.Pulses {
			for _, indicator := range pulse.Indicators {

				// check if type is supported, some OTX indicators don't 
				// translate to the intel framework.
				if iocType, ok := mapOtxType[*indicator.Type]; ok {
					// update the url to be specific to the pulse
					meta.URL = fmt.Sprintf(otxPulseURI, *pulse.ID)

					// build intelligence framework entry
					ioc := brointel.Item{
						*indicator.Indicator,
						iocType,
						meta,
					}

					// print intelligence framework entry
					fmt.Println(ioc.String())
				}
			}
		}
	}
}