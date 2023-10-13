package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Snawoot/drng"
)

const (
	ProgName = "drng"
)

type urlList []string

func (l *urlList) String() string {
	if l == nil {
		return ""
	}
	return strings.Join(*l, " ")
}

func (l *urlList) Set(arg string) error {
	elems := strings.FieldsFunc(arg, func(c rune) bool {
		return c == ' '
	})
	*l = elems
	return nil
}

type byteSliceArg []byte

func (s *byteSliceArg) String() string {
	return hex.EncodeToString(*s)
}

func (s *byteSliceArg) Set(arg string) error {
	b, err := hex.DecodeString(arg)
	if err != nil {
		return err
	}
	*s = b
	return nil
}

var (
	version = "undefined"

	timeout = flag.Duration("timeout", 10*time.Second, "network operation timeout")
	urls    = urlList{
		"https://api.drand.sh",
		"https://drand.cloudflare.com",
	}
	chainHash = byteSliceArg(drng.Must[[]byte](hex.DecodeString("8990e7a9aaed2ffed73dbd7092123d6f289930540d7651336225dc172e51b2ce")))
)

func init() {
	flag.Var(&urls, "api-urls", "list of drand HTTP API URLs separated by space")
	flag.Var(&chainHash, "chainhash", "trust root of chain and reference to chain parameters")
}

func cmdChoice(variants ...string) int {
	fmt.Println(urls)
	fmt.Printf("%x\n", chainHash)
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

func must[Value any](value Value, err error) Value {
	if err != nil {
		panic(err)
	}
	return value
}
