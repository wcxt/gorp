package gorp

import (
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
)

const ProxyPseudonym = "gorp"

// https://datatracker.ietf.org/doc/html/rfc2616#section-13.5.1
var hopByHopHeaders = []string{
	"Transfer-Encoding",
	"TE",
	"Connection",
	"Keep-Alive",
	"Upgrade",
	"Trailer",
	"Proxy-Authorization",
	"Proxy-Authenticate",
}

type ReverseProxy struct {
	Upstream        *url.URL
	RemoveXFHeaders bool
	AddXFHeaders    bool
}

func (rp *ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := r.Clone(context.Background())

	removeHopByHopHeaders(req.Header)

	if rp.RemoveXFHeaders {
		req.Header.Del("X-Forwarded-For")
		req.Header.Del("X-Forwarded-Host")
		req.Header.Del("X-Forwarded-Proto")
	}
	if rp.AddXFHeaders {
		addXForwardedHeaders(req)
	}

	req.URL.Host = rp.Upstream.Host
	req.URL.Scheme = rp.Upstream.Scheme
	req.Host = rp.Upstream.Host

	prior := req.Header.Get("Via")
	if prior != "" {
		req.Header.Set("Via", prior+", "+r.Proto+" "+ProxyPseudonym)
	} else {
		req.Header.Set("Via", r.Proto+" "+ProxyPseudonym)
	}

	// downstream clients should accept trailer fields
	// See https://datatracker.ietf.org/doc/html/rfc9110#name-limitations-on-use-of-trail
	TEValues := strings.Split(r.Header.Get("TE"), ",")
	for _, val := range TEValues {
		if val == "trailers" {
			req.Header.Set("TE", "trailers")
		}
	}

	res, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	defer res.Body.Close()

	removeHopByHopHeaders(res.Header)

	// Headers Must be set before Write or WriteHeader
	for header, value := range res.Header {
		for _, v := range value {
			w.Header().Add(header, v)
		}
	}

	// Trailer Support: Set Trailers header
	trailers := make([]string, len(res.Trailer))
	for trailer, _ := range res.Trailer {
		trailers = append(trailers, trailer)
	}
	w.Header().Add("Trailer", strings.Join(trailers, ", "))

	w.WriteHeader(res.StatusCode)

	// Buffered copy from body to writer
	if _, err := io.Copy(w, res.Body); err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	// Trailer Support: Set trailers
	for trailer, value := range res.Trailer {
		for _, v := range value {
			w.Header().Add(trailer, v)
		}
	}
}

func removeHopByHopHeaders(h http.Header) {
	// handle hop-by-hop headers specified by client
	headers := strings.Split(h.Get("Connection"), ",")
	for _, header := range headers {
		h.Del(strings.TrimSpace(header))
	}

	for _, header := range hopByHopHeaders {
		h.Del(header)
	}
}

func addXForwardedHeaders(r *http.Request) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		panic(err)
	}

	prior := r.Header.Get("X-Forwarded-For")
	if prior != "" {
		r.Header.Set("X-Forwarded-For", prior+", "+ip)
	} else {
		r.Header.Set("X-Forwarded-For", ip)
	}

	r.Header.Set("X-Forwarded-Host", r.Host)

	if r.TLS != nil {
		r.Header.Set("X-Forwarded-Proto", "https")
	} else {
		r.Header.Set("X-Forwarded-Proto", "http")
	}
}
