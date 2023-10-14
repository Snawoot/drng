# drng

Distributed, decentralized, deterministic RNG with verifiable output.

drng uses public entropy from [Distributed Randomness Beacon](https://drand.love/)
in order to initialize high-quality random number generator. It allows to
reproduce results of previous runs any time as long as round number or time
was recorded.

E.g. you may run it to make a random choice from some options:

```
$ drng choice pizza pasta salad soup
Round: 3393047
Round time: 2023-10-13T21:40:30+03:00
Choice: soup
```

Anyone now can verify you've got this outcome in that round:

```
$ drng --round 3393047 choice pasta pizza soup salad
Round: 3393047
Round time: 2023-10-13T21:40:30+03:00
Choice: soup
```

## Installation

### Binaries

Pre-built binaries are available [here](https://github.com/Snawoot/drng/releases/latest).

### Build from source

Alternatively, you may install drng from source. Run the following command within the source directory:

```
make install
```

## Synopsis

```
$ drng -h
Usage:

drng [OPTION]... choice VARIANT [VARIANT]...

  Randomly choose one of VARIANTs.

drng [OPTION]... sample SIZE [FILE]

  Sample of lines with size SIZE from stdin or FILE if specified.

drng [OPTION]... float

  Generates random float value 0 <= X < 1

drng [OPTION]... int N

  Generates random integer value 0 <= X < N

drng [OPTION]... stream

  Output infinite stream of random bytes.

drng version

  Output program version and exit.

Options:
  -api-urls value
    	list of drand HTTP API URLs separated by space (default https://api.drand.sh https://drand.cloudflare.com)
  -chainhash value
    	trust root of chain and reference to chain parameters (default 8990e7a9aaed2ffed73dbd7092123d6f289930540d7651336225dc172e51b2ce)
  -hex-seed value
    	override seed with byte array specified by hex-encoded string
  -hkdf-info bytes
    	override default info bytes supplied to HKDF function (default 64726e672073656564207631)
  -round uint
    	use specific round number
  -round-at time
    	find round happened at time, specified in RFC3339 format (e.g. "2006-01-02T15:04:05+07:00")
  -seed string
    	override seed by string
  -timeout duration
    	network operation timeout (default 10s)
```
