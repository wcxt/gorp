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
	TestProto        = "HTTP/1.1"
)

func TestHandleProxyRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.TransferEncoding) != 0 {
			t.Errorf("got Transfer-Encoding = %v, want %v", r.TransferEncoding, []string{})
		}
		if r.Header.Get("TE") != "trailers" {
			t.Errorf("got TE = %s, want %s", r.Header.Get("TE"), "trailers")
		}
		if r.Header.Get("Connection") != "" {
			t.Errorf("got Connection = %s, want %s", r.Header.Get("Connection"), "")
		}
		if r.Header.Get("Keep-Alive") != "" {
			t.Errorf("got Keep-Alive = %s, want %s", r.Header.Get("Keep-Alive"), "")
		}
		if r.Header.Get("Upgrade") != "" {
			t.Errorf("got Upgrade = %s, want %s", r.Header.Get("Upgrade"), "")
		}
		if len(r.Trailer) != 0 {
			t.Errorf("got Trailer = %v, want %v", r.Trailer, http.Header{})
		}
		if r.Header.Get("Proxy-Authorization") != "" {
			t.Errorf("got Proxy-Authorization = %s, want %s", r.Header.Get("Proxy-Authorization"), "")
		}
		if r.Header.Get("Proxy-Authenticate") != "" {
			t.Errorf("got Proxy-Authenticate = %s, want %s", r.Header.Get("Proxy-Authenticate"), "")
		}
		if r.Header.Get("Remove-Header") != "" {
			t.Errorf("got Remove-Header = %s, want %s", r.Header.Get("Remove-Header"), "")
		}

		want := TestProto + " " + gorp.ProxyPseudonym
		if r.Header.Get("Via") != want {
			t.Errorf("got Via = %s, want %s", r.Header.Get("Via"), want)
		}

		w.Header().Set("Response-Header", "test")
		w.Header().Set("Trailer", "Trailer-Header")
		w.WriteHeader(TestStatusCode)
		w.Write([]byte(TestResponseBody))
		w.Header().Set("Trailer-Header", "test")
	}))
	defer ts.Close()

	destURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	handle := gorp.HandleProxyRequest(destURL)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", TestProxyURL, nil)
	req.Proto = TestProto
	req.TransferEncoding = []string{"gzip", "chunked"}
	req.Header.Set("TE", "trailers, deflate")
	req.Header.Set("Connection", "close, Remove-Header")
	req.Header.Set("Keep-Alive", "test")
	req.Header.Set("Upgrade", "test")
	req.Header.Set("Trailer", "Trailer-Header")
	req.Header.Set("Proxy-Authorization", "test")
	req.Header.Set("Proxy-Authenticate", "test")
	req.Header.Set("Remove-Header", "test")

	handle(w, req)
	res := w.Result()

	if res.StatusCode != TestStatusCode {
		t.Errorf("got res.StatusCode = %d, want %d", res.StatusCode, TestStatusCode)
	}
	if res.Header.Get("Response-Header") != "test" {
		t.Errorf("got Test-Header = %s, want %s", res.Header.Get("Response-Header"), "test")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != TestResponseBody {
		t.Errorf("got res.Body = %s, want %s", string(body), TestResponseBody)
	}
	if res.Trailer.Get("Trailer-Header") != "test" {
		t.Errorf("got Trailer-Header = %s, want %s", res.Header.Get("Trailer-Header"), "test")
	}
}
