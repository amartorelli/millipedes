package sitemap

import (
	"sync"

	"github.com/sirupsen/logrus"
)

// MemorySitemap is an in-memory implementation of a sitemap backend
type MemorySitemap struct {
	sitemap    map[string][]string
	sitemapMux *sync.RWMutex
}

// NewMemorySitemap returns a new MemorySitemap
func NewMemorySitemap() *MemorySitemap {
	return &MemorySitemap{
		sitemap:    make(map[string][]string, 0),
		sitemapMux: &sync.RWMutex{},
	}
}

// AddChildren adds new child pages to a url
func (s *MemorySitemap) AddChildren(url string, children []string) {
	s.sitemapMux.Lock()
	s.sitemap[url] = children
	s.sitemapMux.Unlock()
}

// IsURLPresent returns true if the URL has been stored already
func (s *MemorySitemap) IsURLPresent(url string) bool {
	s.sitemapMux.RLock()
	_, ok := s.sitemap[url]
	s.sitemapMux.RUnlock()
	if ok {
		logrus.Infof("%s already present", url)
	}
	return ok
}

// GetSitemap returns a map[string][]string representation of the sitemap
func (s *MemorySitemap) GetSitemap() map[string][]string {
	return s.sitemap
}
