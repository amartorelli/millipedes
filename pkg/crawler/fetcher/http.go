package fetcher

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

// HTTPFetcher is a structure representing a Fetcher that uses a HTTP client
type HTTPFetcher struct {
	client *http.Client
}

// NewHTTPFetcher returns a new HTTPFetcher
func NewHTTPFetcher() *HTTPFetcher {
	return &HTTPFetcher{
		client: &http.Client{
			Timeout: time.Second * 5,
		},
	}
}

// Fetch fetches a url and returns the body
func (f *HTTPFetcher) Fetch(url string) (io.Reader, error) {
	resp, err := f.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching %s: %s", url, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s response code %d", url, resp.StatusCode)
	}
	defer resp.Body.Close()

	// read all so that we can close the body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(body), nil
}
