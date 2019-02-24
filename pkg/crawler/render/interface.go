package render

// Render is the render interface
type Render interface {
	UpdateSitemap(sitemap map[string][]string) error
	Render()
}
