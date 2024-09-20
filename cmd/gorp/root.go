package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/wcxt/gorp"
)

const DefaultPort = "8080"

var Port string
var Source, Destination string
var TLSEnabled bool
var TLSCertificatePath, TLSPrivateKeyPath string

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

		http.HandleFunc(Source, gorp.HandleProxyRequest(parsedDst))

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
