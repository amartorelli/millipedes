package crawler

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"sync"
	"time"

	"github.com/amartorelli/millipedes/pkg/crawler/fetcher"
	"github.com/amartorelli/millipedes/pkg/crawler/parser"
	"github.com/amartorelli/millipedes/pkg/crawler/sitemap"
	"github.com/sirupsen/logrus"
)

// Crawler represents the crawler
type Crawler struct {
	entrypoint  string
	domain      string
	domainRE    *regexp.Regexp
	fetcher     fetcher.Fetcher
	ratelimiter <-chan time.Time
	sitemap     sitemap.Sitemap
	queue       chan string
	queueLen    int
	seenURL     map[string]struct{}
	seenURLMux  *sync.RWMutex
	mustStop    bool
	wg          *sync.WaitGroup
	workers     int
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewCrawler returns a new crawler
func NewCrawler(website string, workers, queueLen, fetchIntervalMs int, fetcher fetcher.Fetcher, sitemap sitemap.Sitemap) (*Crawler, error) {
	u, err := url.Parse(website)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Crawler{
		entrypoint:  website,
		domain:      u.Host,
		domainRE:    regexp.MustCompile(fmt.Sprintf(".*%s$", u.Host)),
		fetcher:     fetcher,
		ratelimiter: time.Tick(time.Duration(fetchIntervalMs) * time.Millisecond),
		sitemap:     sitemap,
		seenURL:     make(map[string]struct{}, 0),
		seenURLMux:  &sync.RWMutex{},
		queue:       make(chan string, queueLen),
		queueLen:    0,
		mustStop:    false,
		wg:          &sync.WaitGroup{},
		workers:     workers,
		ctx:         ctx,
		cancel:      cancel,
	}, nil
}

// isURLSeen checks if a URL has been seen already
func (c *Crawler) isURLSeen(url string) bool {
	c.seenURLMux.RLock()
	_, ok := c.seenURL[url]
	c.seenURLMux.RUnlock()
	return ok
}

// addToSeen sets a URL as seen so that the program won't try to crawl the link again
func (c *Crawler) addToSeen(url string) {
	c.seenURLMux.Lock()
	c.seenURL[url] = struct{}{}
	c.seenURLMux.Unlock()
}

// processURL parses a page and queues links after having filtered them
func (c *Crawler) processURL(url string) error {
	logrus.Debugf("processing %s", url)

	// if present skip
	if c.sitemap.IsURLPresent(url) {
		logrus.Debugf("%s website already fetched", url)
		return nil
	}

	// get website
	page, err := c.fetcher.Fetch(url)
	if err != nil {
		return err
	}

	c.addToSeen(url)

	// extract links and set connections for the analysed url
	if !c.mustStop {
		links := parser.ExtractLinks(page, c.entrypoint)
		c.sitemap.AddChildren(url, links)

		// add links to queue
		c.queueFilteredLinks(links)
	}
	return nil
}

// isSameDomain checks if the URL belongs to the same domain as the entry point
func (c *Crawler) isSameDomain(uri string) bool {
	URL, err := url.Parse(uri)
	if err != nil {
		return false
	}
	return c.domainRE.MatchString(URL.Host)
}

// queueFilteredLinks adds links to the queue after having filtered them. Links must belong to
// the same main domain and they will only be added to the queue if they were never processed before.
func (c *Crawler) queueFilteredLinks(links []string) {
	for _, l := range links {
		if !c.isSameDomain(l) {
			continue
		}
		if !c.isURLSeen(l) {
			select {
			case c.queue <- l:
				c.queueLen++
				logrus.Debugf("queuing %s", l)
				c.addToSeen(l)
			case <-c.ctx.Done():
				logrus.Infof("cancelling %s", l)
				return
			}
		}
	}
}

// crawlQueue iterates over the elements in the queue and processes the URLs
func (c *Crawler) crawlQueue() {
	for l := range c.queue {
		<-c.ratelimiter
		c.queueLen--
		err := c.processURL(l)
		if err != nil {
			logrus.Error(err)
		}
	}
	c.wg.Done()
}

// IsDone returns true if the queue is empty so that the crawler can stop
func (c *Crawler) IsDone() bool {
	return c.queueLen == 0
}

// Start starts multiple concurrent workers
func (c *Crawler) Start() {
	logrus.Info("crawler started...")
	for i := 0; i < c.workers; i++ {
		c.wg.Add(1)
		go c.crawlQueue()
	}
	c.queue <- c.entrypoint
}

// Sitemap returns a structure representing the sitemap
func (c *Crawler) Sitemap() map[string][]string {
	return c.sitemap.GetSitemap()
}

// Shutdown triggers a stop of the workers
func (c *Crawler) Shutdown() {
	c.mustStop = true
	c.cancel()
	time.Sleep(5 * time.Second)
	close(c.queue)
	c.wg.Wait()
}
