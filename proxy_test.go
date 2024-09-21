package gorp_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/wcxt/gorp"
)

const (
	TestProxyURL     = "http://localhost:8080"
	TestStatusCode   = http.StatusAccepted
	TestResponseBody = "Hello World!"

	TestHeaderKey   = "Test-Header"
	TestHeaderValue = "testvalue"
)

func TestHandleProxyRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(TestHeaderKey, TestHeaderValue)
		w.WriteHeader(TestStatusCode)
		w.Write([]byte(TestResponseBody))
	}))
	defer ts.Close()

	destURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	handle := gorp.HandleProxyRequest(destURL)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", TestProxyURL, nil)

	handle(w, req)
	res := w.Result()

	if res.StatusCode != TestStatusCode {
		t.Errorf("got res.StatusCode = %d, want %d", res.StatusCode, TestStatusCode)
	}
	if res.Header.Get(TestHeaderKey) != TestHeaderValue {
		t.Errorf("got Test-Header = %s, want %s", res.Header.Get(TestHeaderKey), TestHeaderValue)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != TestResponseBody {
		t.Errorf("got res.Body = %s, want %s", string(body), TestResponseBody)
	}
}
