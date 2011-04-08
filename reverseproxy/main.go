
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var (
	flagBind   = flag.String("bind", "0.0.0.0:80", "Address to bind to")
	flagConfig = flag.String("config", "reverseproxy.conf", "Config file")
)

func main() {
	fmt.Fprintf(os.Stderr, "GoReverseProxy — 2011 — by Petar Maymounkov, petar@csail.mit.edu\n")
	flag.Parse()
	p, err := NewProxyEasy(*flagBind, *flagConfig)
	if err != nil {
		log.Printf("Problem starting: %s\n", err)
		os.Exit(1)
	}
	fmt.Print(p.ConfigString())
	<-make(chan int)
}
