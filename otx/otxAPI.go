/*
Local bindings for the OTX API. It turns out that their go bindings are 
missing several critical features. It was easier to reimplement baseline
functionality inside this tool.
*/

package main

import (
	"net/http"
	"io/ioutil"
	"encoding/json"
	"time"
	"log"
)

const (
	baseURL = "https://otx.alienvault.com/api/v1"
	subscriptionsPath = "/pulses/subscribed"
	apiHeader = "X-OTX-API-KEY"
)

// pulse represents an OTX Pulse
type pulse struct {
	ID 		*	string 		`json:"id"`
	Author		*string		`json:"author_name"`
	Name 		*string 	`json:"name"`
	Description	*string 	`json:"description,omitempty"`
	CreatedAt	*string		`json:"created,omitempty"`
	ModifiedAt	*string		`json:"modified"`
	References	[]string	`json:"references,omitempty"`
	Tags		[]string	`json:"tags,omitempty"`
	Indicators	[]struct {
		ID			*string		`json:"_id"`
		Indicator 	*string 	`json:"indicator"`
		Type		*string		`json:"type"`
		Description	*string		`json:"description,omitempty"`
	}
	Revision	*float32	`json:"revision,omitempty"`
}

// feed represents an OTX feed
type feed struct {
	Pulses	[]pulse	`json:"results"`
	// Provide page values for paginating through returned results.
	NextPageString	*string 	`json:"next"`
	PrevPageString	*string		`json:"prev"`
	Count			int			`json:"count"`
}

type Client struct {
	// HTTP client used to communicate with OTX
	client 		*http.Client
	apiKey 		string
	BaseURL 	string
}

// NewClient creates and returns a new OTX API client
func NewClient(apiKey string) Client {
	client := Client{
		client: &http.Client{Timeout: time.Second * 10},
		apiKey: apiKey,
		BaseURL: baseURL,
	}
	return client
}

// getSubscription is used internally to get a subscription page with specific
// options.
func (c *Client) getSubscription(uri string, args map[string]string) *feed {

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		panic(err.Error())
	}

	req.Header.Set(apiHeader, c.apiKey)

	q := req.URL.Query()
	for arg, value := range args {
		q.Add(arg, value)
	}

	req.URL.RawQuery = q.Encode()

	rawResponse, err := c.client.Do(req)
	if err != nil {
		panic(err.Error())
	}

	body, err := ioutil.ReadAll(rawResponse.Body)
	if err != nil {
		panic(err.Error())
	}

	var f = new(feed)
	err = json.Unmarshal(body, &f)
	if err != nil {
		panic(err.Error())
	}

	return f

}

// IterPulses is a way to request a specific set of pulses with arguments and then
// iterate over them in a way similar to a python generator object
// iteration model shamelessly stolen:
// https://blog.carlmjohnson.net/post/on-using-go-channels-like-python-generators/
func (c *Client) IterPulses(args map[string]string) <-chan pulse {

	uri := c.BaseURL + subscriptionsPath
	ch := make(chan pulse)

	f := c.getSubscription(uri, args)
	go func() {
		defer close(ch)
		for f.NextPageString != nil {
			for _, pulse := range f.Pulses {
				ch <- pulse
			}
			log.Printf("Requesting next page: %s", *f.NextPageString)
			f = c.getSubscription(*f.NextPageString, make(map[string]string))
		}
	}()

	return ch
}