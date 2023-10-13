package drng

import (
	"encoding/hex"
	"time"
)

// Config is a set of options for complete initialization of RNG using
// entropy from drand service
type Config struct {
	Round     uint64
	RoundAt   time.Time
	URLs      []string
	ChainHash []byte
	HKDFInfo  []byte
}

var (
	DefaultURLs = []string{
		"https://api.drand.sh",
		"https://drand.cloudflare.com",
	}
	DefaultChainHash = Must[[]byte](hex.DecodeString("8990e7a9aaed2ffed73dbd7092123d6f289930540d7651336225dc172e51b2ce"))
	DefaultHKDFInfo  = []byte("drng seed v1")
)

func (cfg *Config) populateDefaults() *Config {
	newCfg := new(Config)
	*newCfg = *cfg
	cfg = newCfg
	if len(cfg.URLs) == 0 {
		cfg.URLs = DefaultURLs
	}
	if cfg.ChainHash == nil {
		cfg.ChainHash = DefaultChainHash
	}
	if cfg.HKDFInfo == nil {
		cfg.HKDFInfo = DefaultHKDFInfo
	}
	return cfg
}
