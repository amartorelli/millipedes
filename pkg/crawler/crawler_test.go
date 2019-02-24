package crawler

import (
	"fmt"
	"testing"
	"time"

	"github.com/amartorelli/millipedes/pkg/crawler/fetcher"
	"github.com/amartorelli/millipedes/pkg/crawler/sitemap"
)

var (
	template = `<html>
	<head><title></title></head>
	<body>%s</body>
	</html>
	`
	nolinks    = fmt.Sprintf(template, "<div></div>")
	onelink    = fmt.Sprintf(template, "<a href='https://example.com/contact-us'></a>")
	twolinks   = fmt.Sprintf(template, "<a href='https://example.com/contact-us'></a><a href='https://example.com/careers'></a>")
	threelinks = fmt.Sprintf(template, "<a href='https://example.com/contact-us'></a><a href='https://example.com/careers'></a><a href='https://example.com/blog'></a>")

	fakeWebsites = map[string][]byte{
		"https://example.com":            []byte(twolinks),
		"https://example.com/contact-us": []byte(twolinks),
		"https://example.com/careers":    []byte(nolinks),
		"https://example.com/map":        []byte(threelinks),
	}
)

func TestIsURLSeen(t *testing.T) {
	c, err := NewCrawler("https://example.com", 1, 10, 200, fetcher.NewMockFetcher(fakeWebsites), sitemap.NewMemorySitemap())
	if err != nil {
		t.Error(err)
	}

	page := "https://example.com/contact-us"
	if c.isURLSeen(page) {
		t.Errorf("expecting %s to be not found", page)
	}
	c.addToSeen("https://example.com/contact-us")

	if !c.isURLSeen(page) {
		t.Errorf("expecting %s to be found", page)
	}
}

func TestProcessURL(t *testing.T) {
	c, err := NewCrawler("https://example.com", 1, 10, 200, fetcher.NewMockFetcher(fakeWebsites), sitemap.NewMemorySitemap())
	if err != nil {
		t.Error(err)
	}

	// testing entry point, expecting two pages
	err = c.processURL("https://example.com")
	if err != nil {
		t.Error(err)
	}
	if c.queueLen != 2 {
		t.Errorf("expecting two pages in https://example.com")
	}

	// if we execute again we expect no more elements in the queue
	err = c.processURL("https://example.com")
	if err != nil {
		t.Error(err)
	}
	if c.queueLen != 2 {
		t.Errorf("expecting two pages in https://example.com because already analysed")
	}

	// if we fetch a page that doesn't exist we expect an error
	err = c.processURL("https://example.com/nothere")
	if err == nil {
		t.Error("expecting https://example.com/nothere to be not found")
	}

	// if mustStop is set we shouldn't ad any more links to the queue
	c.mustStop = true
	err = c.processURL("https://example.com/map")
	if err != nil {
		t.Error(err)
	}
	if c.queueLen != 2 {
		t.Errorf("expecting no more links in the queue because mustStop is set")
	}
}

func TestIsSameDomain(t *testing.T) {
	c, err := NewCrawler("https://example.com", 1, 10, 200, fetcher.NewMockFetcher(fakeWebsites), sitemap.NewMemorySitemap())
	if err != nil {
		t.Error(err)
	}

	tt := []struct {
		uri string
		res bool
	}{
		{"https://example.com", true},
		{"https://community.example.com", true},
		{"https://community.example.com/testpage", true},
		{"https://google.com", false},
		{"https://google.com/example.com", false},
	}

	for _, tc := range tt {
		same := c.isSameDomain(tc.uri)
		if same != tc.res {
			t.Errorf("expecting isSameDomain(%s) with domain %s to be %v, got %v", tc.uri, c.domain, tc.res, same)
		}
	}
}

func TestQueueFilteredLinks(t *testing.T) {
	c, err := NewCrawler("https://example.com", 1, 10, 200, fetcher.NewMockFetcher(fakeWebsites), sitemap.NewMemorySitemap())
	if err != nil {
		t.Error(err)
	}

	// queuing nothing
	c.queueFilteredLinks([]string{})
	if c.queueLen > 0 {
		t.Errorf("expecting the queue length to be 0 because no links were queued")
	}

	// trying to queue a link that doesn't belong to the same domain
	c.queueFilteredLinks([]string{"www.google.com/test"})
	if c.queueLen > 0 {
		t.Errorf("expecting the queue length to be 0 because the link doesn't belong to the same domain")
	}

	// trying to queue a link that doesn't belong to the same domain and one that does
	c.queueFilteredLinks([]string{"www.google.com/test", "http://community.example.com/"})
	if c.queueLen != 1 {
		t.Errorf("expecting the queue length to be 1 because out of the two links pushed only one belongs to the same domain")
	}

	// trying to queue a link that has already been seen
	c.queueFilteredLinks([]string{"http://community.example.com/"})
	if c.queueLen != 1 {
		t.Errorf("expecting the queue length to be 1 because the link we pushed had already been seen before")
	}
}

func TestIsDone(t *testing.T) {
	c, err := NewCrawler("https://example.com", 1, 10, 200, fetcher.NewMockFetcher(fakeWebsites), sitemap.NewMemorySitemap())
	if err != nil {
		t.Error(err)
	}

	// when the queue has at least one link IsDone must return false
	c.queueFilteredLinks([]string{"http://community.example.com/"})
	if c.IsDone() {
		t.Errorf("expecting the crawler to be still active")
	}

	// emptying the queue and ecpecting the crawler to finish
	go c.crawlQueue()
	time.Sleep(1 * time.Second)
	if !c.IsDone() {
		t.Errorf("expecting the crawling to be finished")
	}
}
