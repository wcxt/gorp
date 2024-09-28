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

const DefaultPort = 8080

var Port int
var Source, Destination string
var TLSEnabled bool
var TLSCertificatePath, TLSPrivateKeyPath string

func init() {
	rootCmd.Flags().IntVarP(&Port, "port", "p", DefaultPort, "port at which proxy will be accepting requests")
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
		if err := gorp.ValidatePort(Port); err != nil {
			return fmt.Errorf("port: %w", err)
		}
		if err := gorp.ValidateSourceFlag(Source); err != nil {
			return fmt.Errorf("source: %w", err)
		}
		if err := gorp.ValidateDestinationFlag(Destination); err != nil {
			return fmt.Errorf("destination: %w", err)
		}

		if TLSEnabled {
			_, err := tls.LoadX509KeyPair(TLSCertificatePath, TLSPrivateKeyPath)
			if err != nil {
				return fmt.Errorf("tls: %w", err)
			}
		}

		addr := ":" + strconv.Itoa(Port)
		parsedDst, _ := url.Parse(Destination)

		http.HandleFunc(Source, gorp.HandleRequest(parsedDst))

		if TLSEnabled {
			log.Fatal(http.ListenAndServeTLS(addr, TLSCertificatePath, TLSPrivateKeyPath, nil))
		} else {
			log.Fatal(http.ListenAndServe(addr, nil))
		}

		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
