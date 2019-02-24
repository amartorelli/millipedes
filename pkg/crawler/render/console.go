package render

import (
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
)

// ConsoleRender renders the sitemap printing it out on the console
type ConsoleRender struct {
	sitemap map[string][]string
}

// NewConsoleRender returns a new ConsoleRender
func NewConsoleRender() *ConsoleRender {
	return &ConsoleRender{sitemap: make(map[string][]string, 0)}
}

// Render renders the sitemap
func (r *ConsoleRender) Render() {
	b, err := json.MarshalIndent(r.sitemap, "", "  ")
	if err != nil {
		logrus.Error(err)
		return
	}
	fmt.Print(string(b))
}

// UpdateSitemap updates the sitemap
func (r *ConsoleRender) UpdateSitemap(sitemap map[string][]string) error {
	r.sitemap = sitemap
	return nil
}
