package fetcher

import "io"

// Fetcher is the interface to abstract the fetch of a url
type Fetcher interface {
	Fetch(url string) (io.Reader, error)
}
