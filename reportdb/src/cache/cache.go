package cache

import (
	"fmt"
	"github.com/dgraph-io/ristretto/v2"
	. "reportdb/utils"
)

var globalCache *ristretto.Cache[string, []DataPoint]

func InitCache() error {

	var err error

	globalCache, err = ristretto.NewCache(&ristretto.Config[string, []DataPoint]{
		NumCounters: 1e7,     // Track 10M keys
		MaxCost:     1 << 30, // 1GB max cache size
		BufferItems: 64,      // Number of keys per Get buffer
		Metrics:     true,
	})

	if err != nil {

		return fmt.Errorf("failed to initialize cache: %w", err)
	}

	return nil
}

func GetCache() *ristretto.Cache[string, []DataPoint] {

	return globalCache
}

func GetMetrics() (hits, misses uint64, hitRatio float64) {

	metrics := globalCache.Metrics

	hits = metrics.Hits()

	misses = metrics.Misses()

	if hits+misses > 0 {

		hitRatio = float64(hits) / float64(hits+misses)
	}

	return hits, misses, hitRatio
}
