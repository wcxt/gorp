package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"net/url"
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
		log.Fatalf("required flag -dst was not provided")
	}
	if seenFlags["cert"] != seenFlags["key"] {
		log.Fatalf("if flag -cert or -key is set both flags must be set")
	}

	if err := gorp.ValidatePort(*portFlag); err != nil {
		log.Fatalf("port: %s", err)
	}
	if err := gorp.ValidateSourceFlag(*sourceFlag); err != nil {
		log.Fatalf("source: %s", err)
	}
	if err := gorp.ValidateDestinationFlag(*destinationFlag); err != nil {
		log.Fatalf("destination: %s", err)
	}

	if *certFileFlag != "" || *keyFileFlag != "" {
		_, err := tls.LoadX509KeyPair(*certFileFlag, *keyFileFlag)
		if err != nil {
			log.Fatalf("tls: %s", err)
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
