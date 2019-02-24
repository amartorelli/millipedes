package parser

import (
	"io"
	"net/url"

	"golang.org/x/net/html"
)

// normaliseURL converts to absolute paths
func normaliseURL(base, href string) string {
	uri, err := url.Parse(href)
	if err != nil {
		return ""
	}
	baseURL, err := url.Parse(base)
	if err != nil {
		return ""
	}
	uri = baseURL.ResolveReference(uri)
	return uri.String()
}

// getLinkFromToken extracts the link from a HTML <a><href> node
func getLinkFromToken(t html.Token) (string, bool) {
	if t.Data == "a" {
		for _, attr := range t.Attr {
			if attr.Key == "href" {
				return attr.Val, true
			}
		}
	}
	return "", false
}

// ExtractLinks returns a list of links from the body of a page. It also requires the baseURL so that it can normalise relative links.
func ExtractLinks(body io.Reader, base string) []string {
	links := []string{}
	tokenizer := html.NewTokenizer(body)
	for {
		t := tokenizer.Next()
		switch t {
		case html.ErrorToken:
			return links
		case html.StartTagToken, html.EndTagToken:
			token := tokenizer.Token()
			l, found := getLinkFromToken(token)
			if !found {
				continue
			}
			links = append(links, normaliseURL(base, l))
		}
	}
}
