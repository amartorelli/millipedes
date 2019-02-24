package render

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDataHandler(t *testing.T) {
	s := NewSigmajsRender(":9876")
	sitemap := map[string][]string{
		"https://example.com": []string{"https://example.com", "https://example.com/a"},
	}
	s.UpdateSitemap(sitemap)

	rr := httptest.NewRecorder()
	// when hitting the /data endpoint it should return json content
	req, err := http.NewRequest("GET", "/data", nil)
	if err != nil {
		t.Fatal(err)
	}
	s.dataHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expecting status code %d, received %d", http.StatusOK, rr.Code)
	}

	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expecting content type to be application/json, received %s", ct)
	}

	body, err := ioutil.ReadAll(rr.Body)
	if err != nil {
		t.Error(err)
	}
	if len(body) == 0 {
		t.Error("expecting the JSON payload to have content")
	}

	err = s.server.Close()
	if err != nil {
		t.Error(err)
	}
}
