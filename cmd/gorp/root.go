package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

const DefaultPort = "8080"
const ProxyPseudonym = "gorp"

var Port string
var Source, Destination string
var TLSEnabled bool
var TLSCertificatePath, TLSPrivateKeyPath string

// https://datatracker.ietf.org/doc/html/rfc2616#section-13.5.1
var HopByHopHeaders = []string{
	"Transfer-Encoding",
	"TE",
	"Connection",
	"Keep-Alive",
	"Upgrade",
	"Trailer",
	"Proxy-Authorization",
	"Proxy-Authenticate",
}

func validateOnlyPathURL(unverifiedURL *url.URL) bool {
	return unverifiedURL.Scheme == "" &&
		unverifiedURL.User == nil &&
		unverifiedURL.Host == "" &&
		unverifiedURL.RawQuery == "" &&
		unverifiedURL.Fragment == ""
}

// NOTE: Support only HTTP scheme
func validateOriginServerURL(unverifiedURL *url.URL) bool {
	return unverifiedURL.Scheme == "http" &&
		unverifiedURL.User == nil &&
		unverifiedURL.Host != "" &&
		unverifiedURL.Path == "" &&
		unverifiedURL.RawQuery == "" &&
		unverifiedURL.Fragment == ""
}

func removeHopByHopHeaders(r *http.Request) {
	for _, header := range HopByHopHeaders {
		r.Header.Del(header)
	}

	// handle hop-by-hop headers specified by client
	headers := strings.Split(r.Header.Get("Connection"), ",")
	for _, header := range headers {
		r.Header.Del(strings.TrimSpace(header))
	}
}

func handleProxyRequest(dest *url.URL) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		req := r.Clone(context.Background())

		req.URL.Host = dest.Host
		req.URL.Scheme = dest.Scheme
		req.Host = dest.Host

		removeHopByHopHeaders(req)

		req.Header.Add("Via", r.Proto+" "+ProxyPseudonym)

		res, err := http.DefaultTransport.RoundTrip(req)
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		defer res.Body.Close()

		// Headers Must be set before Write or WriteHeader
		for header, value := range res.Header {
			for _, v := range value {
				w.Header().Add(header, v)
			}
		}

		// Set Trailers header
		for trailer, _ := range res.Trailer {
			w.Header().Add("Trailer", trailer)
		}

		// Buffered copy from body to writer
		if _, err := io.Copy(w, res.Body); err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusBadGateway)
			return
		}

		// Set trailers
		for trailer, value := range res.Trailer {
			for _, v := range value {
				w.Header().Add(trailer, v)
			}
		}

	}
}

func init() {
	rootCmd.Flags().StringVarP(&Port, "port", "p", DefaultPort, "port at which proxy will be accepting requests")
	rootCmd.Flags().StringVar(&Source, "src", "/", "HTTP path to which origin server will be proxied to")
	rootCmd.Flags().StringVar(&Destination, "dst", "", "proxied HTTP origin server (required)")
	rootCmd.Flags().BoolVar(&TLSEnabled, "tls", false, "use TLS protocol for HTTP proxy encryption")
	rootCmd.Flags().StringVar(&TLSCertificatePath, "cert", "", "TLS certificate path used for TLS encryption (required for --tls)")
	rootCmd.Flags().StringVar(&TLSPrivateKeyPath, "key", "", "TLS private key path used for TLS encryption (required for --tls)")
	rootCmd.MarkFlagsRequiredTogether("tls", "cert", "key")
	rootCmd.MarkFlagRequired("dst")
}

var rootCmd = cobra.Command{
	Use:   "gorp",
	Short: "Gorp is a simple single host HTTP reverse proxy server",
	RunE: func(cmd *cobra.Command, args []string) error {
		tlsConfig := &tls.Config{}

		parsedSrc, err := url.Parse(Source)
		if err != nil || !validateOnlyPathURL(parsedSrc) {
			return fmt.Errorf("Invalid --src flag set")
		}

		parsedDst, err := url.Parse(Destination)
		if err != nil || !validateOriginServerURL(parsedDst) {
			return fmt.Errorf("Invalid --dst flag set")
		}

		parsedPort, err := strconv.Atoi(Port)
		if err != nil || parsedPort < 0 {
			return fmt.Errorf("Invalid --port flag set")
		}

		if TLSEnabled {
			cert, err := tls.LoadX509KeyPair(TLSCertificatePath, TLSPrivateKeyPath)
			if err != nil {
				return fmt.Errorf("Invalid --tls configuration set: %w", err)
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}

		http.HandleFunc(Source, handleProxyRequest(parsedDst))

		if TLSEnabled {
			log.Fatal(http.ListenAndServeTLS(":"+Port, TLSCertificatePath, TLSPrivateKeyPath, nil))
		} else {
			log.Fatal(http.ListenAndServe(":"+Port, nil))
		}

		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
