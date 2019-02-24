package parser

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestNormaliseURL(t *testing.T) {
	tt := []struct {
		base string
		href string
		res  string
	}{
		{"https://www.example.com", "https://www.example.com", "https://www.example.com"},
		{"https://www.example.com", "https://www.example.com/contact-us", "https://www.example.com/contact-us"},
		{"https://www.example.com", "contact-us", "https://www.example.com/contact-us"},
		{"https://www.example.com", "/contact-us.html", "https://www.example.com/contact-us.html"},
		{"https://www.example.com", "/contact-us.html?page=2", "https://www.example.com/contact-us.html?page=2"},
		{"https://www.example.com", "/contact-us.html?page=2&id=3", "https://www.example.com/contact-us.html?page=2&id=3"},
	}

	for _, tc := range tt {
		norm := normaliseURL(tc.base, tc.href)
		if norm != tc.res {
			t.Errorf("%s normalised should be %s, got %s", tc.href, tc.res, norm)
		}
	}
}

func TestGetLinkFromToken(t *testing.T) {
	tt := []struct {
		token html.Token
		out   string
		found bool
	}{
		{
			html.Token{
				Data: "a",
				Attr: []html.Attribute{
					{Key: "href", Val: "https://example.com"},
				},
			},
			"https://example.com",
			true,
		},
		{
			html.Token{
				Data: "a",
				Attr: []html.Attribute{
					{Key: "no", Val: "https://example.com"},
				},
			},
			"",
			false,
		},
		{
			html.Token{
				Data: "div",
				Attr: []html.Attribute{
					{Key: "href", Val: "https://example.com"},
				},
			},
			"",
			false,
		},
	}

	for _, tc := range tt {
		l, ok := getLinkFromToken(tc.token)
		if ok != tc.found {
			t.Errorf("expecting link %s found %v, got %v", l, tc.found, ok)
		}
		if l != tc.out {
			t.Errorf("expected link %s, got %s", tc.out, l)
		}
	}
}

var (
	template = `<html>
	<head><title></title></head>
	<body>%s</body>
	</html>
	`
	nolinks  = fmt.Sprintf(template, "<div></div>")
	onelink  = fmt.Sprintf(template, "<a href='https://example.com/contact-us'></a>")
	twolinks = fmt.Sprintf(template, "<a href='https://example.com/contact-us'></a><a href='https://example.com/careers'></a>")
)

func TestExtractLinks(t *testing.T) {
	tt := []struct {
		body  io.Reader
		base  string
		links []string
	}{
		{
			strings.NewReader(nolinks),
			"https://example.com",
			[]string{},
		},
		{
			strings.NewReader(onelink),
			"https://example.com",
			[]string{"https://example.com/contact-us"},
		},
		{
			strings.NewReader(twolinks),
			"https://example.com",
			[]string{"https://example.com/contact-us", "https://example.com/careers"},
		},
	}

	for _, tc := range tt {
		ll := ExtractLinks(tc.body, tc.base)
		if !reflect.DeepEqual(ll, tc.links) {
			t.Errorf("expecting links %v, got %v", tc.links, ll)
		}
	}
}
