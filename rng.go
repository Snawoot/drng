package drng

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"github.com/drand/drand/chain"
	"github.com/drand/drand/client"
	"github.com/drand/drand/client/http"
	"golang.org/x/crypto/hkdf"
	"pgregory.net/rand"
)

const (
	seedUintsNeeded = 3
	seedUintsSize   = 8
)

// ResultInfo contains additional information about seed of RNG
type ResultInfo struct {
	Round uint64
	At    time.Time
}

// New constructs RNG intialized by seed from drand beacon
func New(ctx context.Context, cfg *Config) (*rand.Rand, *ResultInfo, error) {
	if cfg == nil {
		cfg = new(Config)
	}
	cfg = cfg.populateDefaults()

	c, err := client.New(
		client.From(forURLs(ctx, cfg.URLs, cfg.ChainHash)...),
		client.WithChainHash(cfg.ChainHash),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to construct client: %w", err)
	}

	chainInfo, err := c.Info(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to retrieve chain info: %w", err)
	}

	round := cfg.Round
	roundAt := cfg.RoundAt
	if round == 0 && !roundAt.IsZero() {
		round = c.RoundAt(roundAt)
	}

	drandResult, err := c.Get(ctx, round)
	if err != nil {
		return nil, nil, fmt.Errorf("drand request failed: %w", err)
	}

	rng, err := FromSeed(drandResult.Randomness(), cfg.HKDFInfo)
	if err != nil {
		return nil, nil, fmt.Errorf("RNG construction from seed failed: %w", err)
	}

	round = drandResult.Round()
	roundTime := time.Unix(chain.TimeOfRound(chainInfo.Period, chainInfo.GenesisTime, round), 0)

	return rng, &ResultInfo{
		Round: round,
		At:    roundTime,
	}, nil

}

// FromSeed constructs RNG from seed specified by raw bytes
func FromSeed(seed, info []byte) (*rand.Rand, error) {
	kdf := hkdf.New(sha256.New, seed, nil, info)

	var seedBytes [seedUintsNeeded * seedUintsSize]byte
	if _, err := io.ReadFull(kdf, seedBytes[:]); err != nil {
		return nil, fmt.Errorf("KDF stream read failed: %w", err)
	}

	var seedUints [seedUintsNeeded]uint64

	for i := 0; i < seedUintsNeeded; i++ {
		seedUints[i] = binary.BigEndian.Uint64(seedBytes[i*seedUintsSize : (i+1)*seedUintsSize])
	}

	return rand.New(seedUints[:]...), nil
}

func forURLs(ctx context.Context, urls []string, chainHash []byte) []client.Client {
	clients := make([]client.Client, 0)
	var info *chain.Info
	var skipped []string
	for _, u := range urls {
		if info == nil {
			if c, err := http.New(u, chainHash, nil); err == nil {
				// Note: this wrapper assumes the current behavior that if `New` succeeds,
				// Info will have been fetched.
				info, _ = c.Info(ctx)
				clients = append(clients, c)
			} else {
				skipped = append(skipped, u)
			}
		} else {
			if c, err := http.NewWithInfo(u, info, nil); err == nil {
				clients = append(clients, c)
			}
		}
	}
	if info != nil {
		for _, u := range skipped {
			if c, err := http.NewWithInfo(u, info, nil); err == nil {
				clients = append(clients, c)
			}
		}
	}
	return clients
}
