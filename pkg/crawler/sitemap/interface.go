package sitemap

// Sitemap is an interfaces that represents a backend where to store the sitemap
type Sitemap interface {
	AddChildren(url string, children []string)
	IsURLPresent(url string) bool
	GetSitemap() map[string][]string
}
