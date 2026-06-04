package database

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

var (
	globalGenOnce sync.Once
	globalGen     *TSIDGenerator
	globalGenErr  error
)

func GetTSIDGenerator() (*TSIDGenerator, error) {
	globalGenOnce.Do(func() {
		globalGen, globalGenErr = newTSIDGenerator()
	})
	return globalGen, globalGenErr
}

func newTSIDGenerator() (*TSIDGenerator, error) {
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

func (g *TSIDGenerator) Generate() (int64, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	t, err := g.factory.Generate()
	if err != nil {
		return 0, fmt.Errorf("failed to generate TSID: %w", err)
	}

	return t.ToNumber(), nil
}

func Int64ToStringID(id int64) string {
	if id == 0 {
		return "0"
	}

	t := tsid.FromNumber(id)
	return t.ToString()
}

func StringIDToInt64(s string) (int64, error) {
	if s == "" || s == "0" {
		return 0, nil
	}

	var t *tsid.Tsid
	func() {
		defer func() {
			if r := recover(); r != nil {
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

func MustInt64ToStringID(id int64) string {
	if id == 0 {
		return "0"
	}
	return Int64ToStringID(id)
}

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
