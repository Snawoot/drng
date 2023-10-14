package main

import (
	"bufio"
	"context"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Snawoot/drng"
	"github.com/Snawoot/terse/reservoir"
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
	chainHash = byteSliceArg(drng.Must(hex.DecodeString("8990e7a9aaed2ffed73dbd7092123d6f289930540d7651336225dc172e51b2ce")))
	round     = flag.Uint64("round", 0, "use specific round number")
	roundAt   = timeArg{}
	seed      = flag.String("seed", "", "override seed by string")
	hexSeed   = byteSliceArg(nil)
	hkdfInfo  = byteSliceArg(drng.DefaultHKDFInfo)
)

func init() {
	flag.Var(&urls, "api-urls", "list of drand HTTP API URLs separated by space")
	flag.Var(&chainHash, "chainhash", "trust root of chain and reference to chain parameters")
	flag.Var(&roundAt, "round-at", "find round happened at `time`, specified in RFC3339 format (e.g. \"2006-01-02T15:04:05+07:00\")")
	flag.Var(&hexSeed, "hex-seed", "override seed with byte array specified by hex-encoded string")
	flag.Var(&hkdfInfo, "hkdf-info", "override default info `bytes` supplied to HKDF function")
}

type resultInfo []struct {
	key   string
	value string
}

func (info resultInfo) Print() {
	for _, pair := range info {
		fmt.Fprintf(os.Stderr, "%s: %s\n", pair.key, pair.value)
	}
}

func makeRand() (*rand.Rand, resultInfo, error) {
	if *seed != "" || hexSeed != nil {
		if *seed != "" && hexSeed != nil {
			return nil, nil, errors.New("seed and hex seed are mutually exclusive options")
		}
		if *seed != "" {
			hexSeed = byteSliceArg(*seed)
		}
		rng, err := drng.FromSeed(hexSeed, hkdfInfo)
		if err != nil {
			return nil, nil, fmt.Errorf(
				"RNG init failed: %w", err,
			)
		}
		return rng, resultInfo{
			{
				key:   "Seed",
				value: hex.EncodeToString(hexSeed),
			},
		}, nil
	}
	cfg := drng.Config{
		Round:     *round,
		RoundAt:   time.Time(roundAt),
		URLs:      urls,
		ChainHash: chainHash,
	}
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	rng, info, err := drng.New(ctx, &cfg)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"RNG init failed: %w", err,
		)
	}
	return rng, resultInfo{
		{
			key:   "Round",
			value: strconv.FormatUint(info.Round, 10),
		},
		{
			key:   "Round time",
			value: info.At.Format(time.RFC3339),
		},
		{
			key:   "Seed",
			value: hex.EncodeToString(info.Seed),
		},
	}, nil
}

func cmdChoice(variants ...string) int {
	if len(variants) == 0 {
		fmt.Fprintln(os.Stderr, "Need at least one variant argument!")
		usage()
		return 2
	}
	rng, info, err := makeRand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "initialization failed: %v", err)
		return 1
	}
	slices.Sort(variants)
	res := variants[rng.Intn(len(variants))]
	info.Print()
	fmt.Println(res)
	return 0
}

func cmdSample(args ...string) int {
	nargs := len(args)
	if nargs < 1 || nargs > 2 {
		fmt.Fprintln(os.Stderr, "Incorrect number of positional arguments!")
		usage()
		return 2
	}

	size, err := strconv.ParseInt(args[0], 10, 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Can't parse SIZE argument: %v", err)
		usage()
		return 2
	}
	if size < 0 {
		fmt.Fprintln(os.Stderr, "error: negative limit value")
		usage()
		return 2
	}

	var input io.Reader = os.Stdin
	if nargs > 1 {
		f, err := os.Open(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to open input file: %v\n", err)
			return 3
		}
		defer f.Close()
		input = f
	}
	var output io.Writer = os.Stdout

	rng, info, err := makeRand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "initialization failed: %v", err)
		return 1
	}

	r := reservoir.NewReservoir[string](int(size), rng)

	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		if idx := r.AddViaIndex(); idx >= 0 {
			r.Load(idx, scanner.Text())
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "read error: %v\n", err)
	}

	info.Print()
	for _, line := range r.Items() {
		if _, err := fmt.Fprintf(output, "%s\n", line); err != nil {
			fmt.Fprintf(os.Stderr, "write error: %v\n", err)
			return 4
		}
	}

	return 0
}

func cmdFloat(args ...string) int {
	if len(args) != 0 {
		fmt.Fprintln(os.Stderr, "Unexpected number of arguments.")
		usage()
		return 2
	}

	rng, info, err := makeRand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "initialization failed: %v", err)
		return 1
	}

	info.Print()
	fmt.Printf("%f\n", rng.Float64())
	return 0
}

func cmdInt(args ...string) int {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "Unexpected number of arguments.")
		usage()
		return 2
	}

	limit, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Can't parse limit: %v", err)
		usage()
		return 2
	}
	if limit == 0 {
		fmt.Fprintln(os.Stderr, "Limit can't be zero!")
		usage()
		return 2
	}

	rng, info, err := makeRand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "initialization failed: %v", err)
		return 1
	}

	info.Print()
	fmt.Printf("%d\n", rng.Uint64n(limit))
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
	fmt.Fprintf(out, "%s [OPTION]... sample SIZE [FILE]\n", ProgName)
	fmt.Fprintf(out, "%s [OPTION]... float\n", ProgName)
	fmt.Fprintf(out, "%s [OPTION]... int N\n", ProgName)
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
	case "sample":
		return cmdSample(args[1:]...)
	case "float":
		return cmdFloat(args[1:]...)
	case "int":
		return cmdInt(args[1:]...)
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
