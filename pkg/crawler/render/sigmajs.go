package render

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

// SigmajsRender renders the sitemap using Sigmajs
type SigmajsRender struct {
	server  *http.Server
	mux     *http.ServeMux
	content []byte
	sitemap map[string][]string
}

// NewSigmajsRender returns a new ConsoleRender
func NewSigmajsRender(lAddr string) *SigmajsRender {
	mux := http.NewServeMux()
	return &SigmajsRender{
		server: &http.Server{
			Addr:           lAddr,
			Handler:        mux,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		},
		mux:     mux,
		sitemap: make(map[string][]string, 0),
	}
}

// sigmaNode represents a sigma node
type sigmaNode struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	X     int    `json:"x"`
	Y     int    `json:"y"`
	Size  int    `json:"size"`
}

// sigmaEdge represents a sigma edge
type sigmaEdge struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
}

// sigma is the main sigma object
type sigma struct {
	Nodes []sigmaNode `json:"nodes"`
	Edges []sigmaEdge `json:"edges"`
}

// sitemapToSigma converts a map[string][]string to a sigma structure
// to allow it to be parsed and visualised by Sigma
func sitemapToSigma(sitemap map[string][]string) sigma {
	s := sigma{
		Nodes: make([]sigmaNode, 0),
		Edges: make([]sigmaEdge, 0),
	}
	allNodes := make(map[string]int, 0)
	seenEdges := make(map[string]struct{}, 0)
	for n, ee := range sitemap {
		allNodes[n]++
		for _, e := range ee {
			allNodes[e]++
			ID := fmt.Sprintf("%s-%s", n, e)

			// don't create edge if already there
			_, ok := seenEdges[ID]
			if ok {
				continue
			}
			seenEdges[ID] = struct{}{}
			se := sigmaEdge{ID: ID, Source: n, Target: e}
			s.Edges = append(s.Edges, se)
		}
	}

	for n := range allNodes {
		s.Nodes = append(s.Nodes, sigmaNode{ID: n, Label: n, X: rand.Intn(100), Y: rand.Intn(100), Size: allNodes[n]})
	}

	return s
}

// openBrowser opens a browser pointing to a URL. This function ensures compatibility with multiple OS
func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

// dataHandler returns the sigma-ready JSON content
func (r *SigmajsRender) dataHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	w.Write(r.content)
}

// Serve starts a HTTP server to serve the graph and the JSON data used by SigmaJS
func (r *SigmajsRender) Serve() error {
	fs := http.FileServer(http.Dir("./static"))
	r.mux.Handle("/", fs)
	r.mux.HandleFunc("/data", r.dataHandler)

	logrus.Info("Starting sigma...")
	err := r.server.ListenAndServe()
	if err != nil {
		logrus.Fatal(err)
	}

	return nil
}

// Render renders the sitemap
func (r *SigmajsRender) Render() error {
	time.Sleep(1 * time.Second)
	go r.Serve()
	time.Sleep(5 * time.Second)
	err := openBrowser("http://localhost:9876/index.htm")
	if err != nil {
		return err
	}
	time.Sleep(5 * time.Second)

	err = r.server.Close()
	if err != nil {
		return err
	}
	return nil
}

// UpdateSitemap updates the sitemap
func (r *SigmajsRender) UpdateSitemap(sitemap map[string][]string) error {
	r.sitemap = sitemap
	sigma := sitemapToSigma(r.sitemap)
	content, err := json.Marshal(sigma)
	if err != nil {
		return err
	}
	r.content = content
	return nil
}
