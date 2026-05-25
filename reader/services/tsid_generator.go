package services

import (
	"fmt"
	"hash/fnv"
	"sync"

	"github.com/rushysloth/go-tsid"
)

type TSIDGenerator struct {
	factory *tsid.TsidFactory
	mu      sync.Mutex
}

// NewTSIDGenerator creates a new TSID generator with default settings
// Single-instance deployment, node bits = 0
func NewTSIDGenerator() (*TSIDGenerator, error) {
	// Create factory with default settings:
	// - Node bits: 0 (single instance)
	// - Default epoch: 2020-01-01
	factory, err := tsid.TsidFactoryBuilder().
		WithNodeBits(0).
		Build()

	if err != nil {
		return nil, fmt.Errorf("failed to build TSID factory: %w", err)
	}

	return &TSIDGenerator{
		factory: factory,
	}, nil
}

// Generate generates a new TSID and returns it as an int64
func (g *TSIDGenerator) Generate() (int64, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	t, err := g.factory.Generate()
	if err != nil {
		return 0, fmt.Errorf("failed to generate TSID: %w", err)
	}

	return t.ToNumber(), nil
}

// Int64ToStringID converts a TSID int64 to its 13-character base32 string representation
func Int64ToStringID(id int64) string {
	if id == 0 {
		return "0"
	}

	// Create TSID from int64 and get string representation
	t := tsid.FromNumber(id)
	return t.ToString()
}

// StringIDToInt64 converts a 13-character base32 string TSID back to int64
func StringIDToInt64(s string) (int64, error) {
	if s == "" || s == "0" {
		return 0, nil
	}

	// Parse the string TSID - FromString may panic on invalid input
	var t *tsid.Tsid
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Recover from panic - this means invalid TSID string
				t = nil
			}
		}()
		t = tsid.FromString(s)
	}()

	if t == nil {
		return 0, fmt.Errorf("failed to parse TSID string '%s'", s)
	}

	return t.ToNumber(), nil
}

// MustInt64ToStringID is like Int64ToStringID but panics on error (useful for models)
func MustInt64ToStringID(id int64) string {
	if id == 0 {
		return "0"
	}
	return Int64ToStringID(id)
}

// HashEntityID returns a deterministic int64 by hashing the given components with FNV-64a.
// Components are joined with ":" as separator. The sign bit is cleared to keep IDs positive.
func HashEntityID(components ...string) int64 {
	h := fnv.New64a()
	for i, c := range components {
		if i > 0 {
			h.Write([]byte(":"))
		}
		h.Write([]byte(c))
	}
	return int64(h.Sum64() >> 1)
}
