package requesthandler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"request"
	"strings"
	"testing"
)

func TestContentType(t *testing.T) {
	var h http.Request
	h.Header = map[string][]string{
		"Content-Type": {"application/json", "text/utf-8"},
	}
	result := checkContentType(&h, httptest.NewRecorder(), "application/json")
	if !result {
		t.Errorf("Expected true, actual: %v", result)
	}
}

func TestContentTypeFail(t *testing.T) {
	var h http.Request
	h.Header = map[string][]string{
		"Content-Type": {"sth-else", "text/utf-8"},
	}
	w := httptest.NewRecorder()
	result := checkContentType(&h, w, "application/json")
	if result {
		t.Errorf("Expected false, actual: %v", result)
	}
	if w.Code != http.StatusUnsupportedMediaType {
		t.Errorf("Expected 415, actual %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), http.StatusText(http.StatusUnsupportedMediaType)) {
		t.Errorf("Expected to contain %s, actual %s", http.StatusText(http.StatusUnsupportedMediaType), w.Body.String())
	}
}

func TestCoordinatesValid(t *testing.T) {
	r := map[*request.PlaceOrderRequest]bool{
		createRequest("-90", "-180", "90", "180"):        true,  // normal case
		createRequest("-90.001", "-180", "90", "180"):    false, // overflow
		createRequest("-90", "-180.001", "90", "180"):    false,
		createRequest("-90", "-180", "90.000001", "180"): false,
		createRequest("-90", "-180", "90", "180.00001"):  false,
		createRequest("-90asdf", "-180", "90", "180"):    false, // invalid number
	}

	for k, v := range r {
		if coordinatesValid(k) != v {
			t.Errorf("Validate coordinate %#v returns %v, expects %v", k, !v, v)
		}
	}
}

func createRequest(oLat string, oLong string, dLat string, dLong string) *request.PlaceOrderRequest {
	return &request.PlaceOrderRequest{
		Origin:      []string{oLat, oLong},
		Destination: []string{dLat, dLong},
	}
}

func TestGetPageAndLimit(t *testing.T) {
	var h http.Request

	r := map[string]*testPageLimitResult{
		"":                 createPageLimitResult(1, -1, 0), //default
		"page=2":           createPageLimitResult(2, -1, 0),
		"limit=2":          createPageLimitResult(1, 2, 0),
		"page=2&limit=2":   createPageLimitResult(2, 2, 0),
		"page=0&limit=2":   createPageLimitResult(1, 2, 1),
		"page=2&limit=-2":  createPageLimitResult(2, -1, 1),
		"page=-2&limit=-2": createPageLimitResult(1, -1, 2),
		"page=a&limit=-2":  createPageLimitResult(1, -1, 2),
	}

	for k, v := range r {
		h.URL = &url.URL{RawQuery: k}
		p, l, e := getPageAndLimit(&h)
		if p != v.Page || l != v.Limit || len(e) != v.NumErr {
			t.Errorf("Get page/limit expects %#v, actual: %d %d %d", v, p, l, len(e))
		}
	}

}

type testPageLimitResult struct {
	Page   int
	Limit  int
	NumErr int
}

func createPageLimitResult(p int, l int, numError int) *testPageLimitResult {
	return &testPageLimitResult{p, l, numError}
}
