package main

import (
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/wcxt/gorp"
)

// NOTE: Maybe this could go to some other file
type FlagStringMap map[string]string

func (f *FlagStringMap) String() string {
	return fmt.Sprintf("%v", *f)
}

func (f *FlagStringMap) Set(value string) error {
	substrings := strings.Split(value, ":")
	if len(substrings) != 2 {
		return errors.New("parse error")
	}
	(*f)[substrings[0]] = substrings[1]
	return nil
}

// TODO: Should probably be moved to go main with the init flags
var (
	AddInHeadersFlag  = FlagStringMap{}
	AddOutHeadersFlag = FlagStringMap{}

	portFlag     = flag.Int("port", 8080, "port at which proxy will be accepting requests")
	pathFlag     = flag.String("path", "/", "HTTP path to which origin server will be proxied to")
	upstreamFlag = flag.String("upstream", "", "proxied HTTP origin server (required)")
	certFileFlag = flag.String("cert", "", "TLS certificate path used for TLS encryption")
	keyFileFlag  = flag.String("key", "", "TLS private key path used for TLS encryption")
	AddXFFlag    = flag.Bool("add-xf-headers", true, "Adds XF headers proxy server request")
	RemoveXFFlag = flag.Bool("remove-xf-headers", true, "Removes XF headers from incoming request")
)

func init() {
	// NOTE: Maybe header list should also be validated
	flag.Var(&AddInHeadersFlag, "add-in-headers", "Adds headers to incoming request")
	flag.Var(&AddOutHeadersFlag, "add-out-headers", "Adds headers to outgoing request")
}

func validate() error {
	if err := gorp.ValidatePort(*portFlag); err != nil {
		return fmt.Errorf("invalid value %d for port: %w", *portFlag, err)
	}
	if err := gorp.ValidatePath(*pathFlag); err != nil {
		return fmt.Errorf("invalid value %s for path: %w", *pathFlag, err)
	}
	if err := gorp.ValidateUpstream(*upstreamFlag); err != nil {
		return fmt.Errorf("invalid value %s for upstream: %w", *upstreamFlag, err)
	}
	if *certFileFlag != "" || *keyFileFlag != "" {
		if _, err := tls.LoadX509KeyPair(*certFileFlag, *keyFileFlag); err != nil {
			return fmt.Errorf("invalid files %s, %s for key pair: %w", *certFileFlag, *keyFileFlag, err)
		}
	}

	return nil
}

func usage() {
	fmt.Fprintln(flag.CommandLine.Output(), "Usage: gorp [flags]\n")
	fmt.Fprintln(flag.CommandLine.Output(), "Flags:")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	seenFlags := map[string]bool{}
	flag.Visit(func(f *flag.Flag) { seenFlags[f.Name] = true })

	if !seenFlags["upstream"] {
		fmt.Fprintln(os.Stderr, "required flag -upstream was not provided")
		flag.Usage()
		os.Exit(2)
	}
	if seenFlags["cert"] != seenFlags["key"] {
		fmt.Fprintln(os.Stderr, "if flag -cert or -key is set both flags must be set")
		flag.Usage()
		os.Exit(2)
	}

	if err := validate(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		flag.Usage()
		os.Exit(2)
	}

	addr := ":" + strconv.Itoa(*portFlag)
	parsedUpstream, err := url.Parse(*upstreamFlag)
	if err != nil {
		panic(err)
	}

	http.Handle(*pathFlag, &gorp.ReverseProxy{
		Upstream:        parsedUpstream,
		AddXFHeaders:    *AddXFFlag,
		RemoveXFHeaders: *RemoveXFFlag,
		AddInHeaders:    AddInHeadersFlag,
		AddOutHeaders:   AddOutHeadersFlag,
	})

	if *certFileFlag != "" || *keyFileFlag != "" {
		log.Fatal(http.ListenAndServeTLS(addr, *certFileFlag, *keyFileFlag, nil))
	} else {
		log.Fatal(http.ListenAndServe(addr, nil))
	}
}
