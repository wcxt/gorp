package gorp_test

import (
	"io"
	"net"
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

func TestHandleRequest(t *testing.T) {
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

	req.Header.Set("X-Forwarded-For", "1.1.1.1")
	req.Host = "example.com"

	req.Header.Set("Via", "john")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.TransferEncoding) != 0 {
			t.Errorf("got Transfer-Encoding = %v, want %v", r.TransferEncoding, []string{})
		}
		if r.Header.Get("TE") != "trailers" {
			t.Errorf("got TE = %s, want %s", r.Header.Get("TE"), "trailers")
		}
		if r.Header.Get("Connection") != "" {
			t.Errorf("got request Connection = %s, want %s", r.Header.Get("Connection"), "")
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
			t.Errorf("got request Remove-Header = %s, want %s", r.Header.Get("Remove-Header"), "")
		}

		want := "john, " + TestProto + " " + gorp.ProxyPseudonym
		if r.Header.Get("Via") != want {
			t.Errorf("got Via = %s, want %s", r.Header.Get("Via"), want)
		}

		host, _, err := net.SplitHostPort(req.RemoteAddr)
		if err != nil {
			t.Fatal(err)
		}

		if r.Header.Get("X-Forwarded-For") != host {
			t.Errorf("got X-Forwarded-For = %s, want %s", r.Header.Get("X-Forwarded-For"), host)
		}
		if r.Header.Get("X-Forwarded-Host") != "example.com" {
			t.Errorf("got X-Forwarded-Host = %s, want %s", r.Header.Get("X-Forwarded-Host"), "example.com")
		}
		if r.Header.Get("X-Forwarded-Proto") != "http" {
			t.Errorf("got X-Forwarded-Proto = %s, want %s", r.Header.Get("X-Forwarded-Proto"), "http")
		}

		if r.Header.Get("Header-In-Config") != "test" {
			t.Errorf("got Header-In-Config = %s, want %s", r.Header.Get("Header-In-Config"), "test")
		}

		w.Header().Set("Response-Header", "test")

		w.Header().Set("Connection", "close, Remove-Header")
		w.Header().Set("Remove-Header", "test")

		w.Header().Set("Trailer", "Trailer-Header, Another-Header")
		w.WriteHeader(TestStatusCode)
		w.Write([]byte(TestResponseBody))
		w.Header().Set("Trailer-Header", "test")
		w.Header().Set("Another-Header", "test")
	}))
	defer ts.Close()

	upstreamURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	handler := gorp.ReverseProxy{
		Upstream:        upstreamURL,
		AddXFHeaders:    true,
		RemoveXFHeaders: true,
		AddInHeaders:    map[string]string{"Header-In-Config": "test"},
		AddOutHeaders:   map[string]string{"Header-Out-Config": "test"},
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	res := w.Result()

	if res.StatusCode != TestStatusCode {
		t.Errorf("got res.StatusCode = %d, want %d", res.StatusCode, TestStatusCode)
	}
	if res.Header.Get("Response-Header") != "test" {
		t.Errorf("got Test-Header = %s, want %s", res.Header.Get("Response-Header"), "test")
	}
	if res.Header.Get("Connection") != "" {
		t.Errorf("got response Connection = %s, want %s", res.Header.Get("Connection"), "")
	}
	if res.Header.Get("Remove-Header") != "test" {
		t.Errorf("got response Remove-Header = %s, want %s", res.Header.Get("Remove-Header"), "")
	}

	if res.Header.Get("Header-Out-Config") != "test" {
		t.Errorf("got Header-Out-Config = %s, want %s", res.Header.Get("Header-Out-Config"), "test")
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
