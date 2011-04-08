
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var (
	flagAddr   = flag.String("addr", "0.0.0.0:80", "Address to bind to")
	flagConfig = flag.String("conf", "reverseproxy.conf", "Config file")
)

func main() {
	fmt.Fprintf(os.Stderr, "GoFrontline — 2011 — by Petar Maymounkov, petar@csail.mit.edu\n")
	flag.Parse()
	p, err := NewProxyEasy(*flagAddr, *flagConfig)
	if err != nil {
		log.Printf("Problem starting: %s\n", err)
		os.Exit(1)
	}
	fmt.Print(p.ConfigString())
	<-make(chan int)
}
