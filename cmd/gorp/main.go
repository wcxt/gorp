package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/wcxt/gorp"
)

var (
	portFlag        = flag.Int("port", 8080, "port at which proxy will be accepting requests")
	sourceFlag      = flag.String("src", "/", "HTTP path to which origin server will be proxied to")
	destinationFlag = flag.String("dst", "", "proxied HTTP origin server (required)")
	certFileFlag    = flag.String("cert", "", "TLS certificate path used for TLS encryption")
	keyFileFlag     = flag.String("key", "", "TLS private key path used for TLS encryption")
)

func validate() error {
	if err := gorp.ValidatePort(*portFlag); err != nil {
		return fmt.Errorf("invalid value %d for port: %w", *portFlag, err)
	}
	if err := gorp.ValidateSourceFlag(*sourceFlag); err != nil {
		return fmt.Errorf("invalid value %s for source: %w", *sourceFlag, err)
	}
	if err := gorp.ValidateDestinationFlag(*destinationFlag); err != nil {
		return fmt.Errorf("invalid value %s for destination: %w", *destinationFlag, err)
	}
	if *certFileFlag != "" || *keyFileFlag != "" {
		if _, err := tls.LoadX509KeyPair(*certFileFlag, *keyFileFlag); err != nil {
			return fmt.Errorf("invalid files %s, %s for key pair: %w", *certFileFlag, *keyFileFlag, err)
		}
	}

	return nil
}

func main() {
	flag.Parse()

	seenFlags := map[string]bool{}
	flag.Visit(func(f *flag.Flag) { seenFlags[f.Name] = true })

	if !seenFlags["dst"] {
		fmt.Fprintln(os.Stderr, "required flag -dst was not provided")
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
	parsedDst, _ := url.Parse(*destinationFlag)

	http.Handle(*sourceFlag, &gorp.ReverseProxy{Destination: parsedDst})

	if *certFileFlag != "" || *keyFileFlag != "" {
		log.Fatal(http.ListenAndServeTLS(addr, *certFileFlag, *keyFileFlag, nil))
	} else {
		log.Fatal(http.ListenAndServe(addr, nil))
	}
}
