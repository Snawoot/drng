package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"time"
)

const (
	ProgName = "drng"
)

var urls = []string{
	"https://api.drand.sh",
	"https://drand.cloudflare.com",
}

var (
	version = "undefined"

	timeout = flag.Duration("timeout", 10*time.Second, "network operation timeout")
)

var chainHash, _ = hex.DecodeString("8990e7a9aaed2ffed73dbd7092123d6f289930540d7651336225dc172e51b2ce")

// func main() {
// 	c, err := client.New(
// 		client.From(http.ForURLs(urls, chainHash)...),
// 		client.WithChainHash(chainHash),
// 	)
// 	if err != nil {
// 		log.Fatalf("can't initialize drand client: %v", err)
// 	}
//
// 	round := c.RoundAt(time.Now())
// 	log.Printf("target round = %d", round)
// 	res, err := c.Get(context.TODO(), round)
// 	if err != nil {
// 		log.Fatalf("request failed: %v", err)
// 	}
//
// 	log.Printf("result: round: %d, randomness: %x, signature: %x", res.Round(), res.Randomness(), res.Signature())
//
// 	res, err = c.Get(context.TODO(), 0)
// 	if err != nil {
// 		log.Fatalf("request failed: %v", err)
// 	}
//
// 	log.Printf("result: round: %d, randomness: %x, signature: %x", res.Round(), res.Randomness(), res.Signature())
// }

func cmdChoice(variants ...string) int {
	return 0
}

func cmdVersion() int {
	fmt.Println(version)
	return 0
}

func usage() {
	out := flag.CommandLine.Output()
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out)
	fmt.Fprintf(out, "%s [OPTION]... choice VARIANT [VARIANT]...\n", ProgName)
	fmt.Fprintf(out, "%s version\n", ProgName)
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Options:")
	flag.PrintDefaults()
}

func run() int {
	flag.CommandLine.Usage = usage
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		// no subcommand
		usage()
		return 2
	}

	switch args[0] {
	case "choice":
		return cmdChoice(args[1:]...)
	case "version":
		return cmdVersion()
	}
	usage()
	return 2
}

func main() {
	os.Exit(run())
}
