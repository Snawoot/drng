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

