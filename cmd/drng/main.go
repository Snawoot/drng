package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Snawoot/drng"
	"pgregory.net/rand"
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

type timeArg time.Time

func (t *timeArg) String() string {
	if t == nil {
		return ""
	}
	return time.Time(*t).Format(time.RFC3339)
}

func (t *timeArg) Set(arg string) error {
	parsed, err := time.Parse(time.RFC3339, arg)
	if err != nil {
		return err
	}
	*t = timeArg(parsed)
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
	round     = flag.Uint64("round", 0, "use specific round number")
	roundAt   = timeArg{}
)

func init() {
	flag.Var(&urls, "api-urls", "list of drand HTTP API URLs separated by space")
	flag.Var(&chainHash, "chainhash", "trust root of chain and reference to chain parameters")
	flag.Var(&roundAt, "round-at", "find round happened at time, specified in RFC3339 format (e.g. \"2006-01-02T15:04:05+07:00\")")
}

func makeRand() (*rand.Rand, *drng.ResultInfo, error) {
	cfg := drng.Config{
		Round:     *round,
		RoundAt:   time.Time(roundAt),
		URLs:      urls,
		ChainHash: chainHash,
	}
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	return drng.New(ctx, &cfg)
}

func cmdChoice(variants ...string) int {
	_, info, err := makeRand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "initialization failed: %v", err)
		return 1
	}
	fmt.Printf("Round: %d\n", info.Round)
	fmt.Printf("Round time: %s\n", info.At.Format(time.RFC3339))
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
