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

	if err := gorp.ValidatePort(*portFlag); err != nil {
		fmt.Fprintf(os.Stderr, "port: %s\n", err)
		flag.Usage()
		os.Exit(2)
	}
	if err := gorp.ValidateSourceFlag(*sourceFlag); err != nil {
		fmt.Fprintf(os.Stderr, "source: %s\n", err)
		flag.Usage()
		os.Exit(2)
	}
	if err := gorp.ValidateDestinationFlag(*destinationFlag); err != nil {
		fmt.Fprintf(os.Stderr, "destination: %s\n", err)
		flag.Usage()
		os.Exit(2)
	}

	if *certFileFlag != "" || *keyFileFlag != "" {
		_, err := tls.LoadX509KeyPair(*certFileFlag, *keyFileFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "tls: %s\n", err)
			flag.Usage()
			os.Exit(2)
		}
	}

	addr := ":" + strconv.Itoa(*portFlag)
	parsedDst, _ := url.Parse(*destinationFlag)

	http.HandleFunc(*sourceFlag, gorp.HandleRequest(parsedDst))

	if *certFileFlag != "" || *keyFileFlag != "" {
		log.Fatal(http.ListenAndServeTLS(addr, *certFileFlag, *keyFileFlag, nil))
	} else {
		log.Fatal(http.ListenAndServe(addr, nil))
	}
}
