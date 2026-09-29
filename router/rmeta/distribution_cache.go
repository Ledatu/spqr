package rmeta

import (
	"time"

	"github.com/pg-sharding/spqr/pkg/models/distributions"
	"github.com/pg-sharding/spqr/router/rfqn"
)

// Distributions change only through metadata DDL; the expiry bounds how long a long-lived connection can route on a stale one.
const distributionCacheTTL = 30 * time.Second

type DistributionCache struct {
	entries  map[rfqn.RelationFQN]*distributions.Distribution
	loadedAt time.Time
}

type DistributionCacheHolder interface {
	RelationDistributionCache() *DistributionCache
}

func NewDistributionCache() *DistributionCache {
	return &DistributionCache{entries: map[rfqn.RelationFQN]*distributions.Distribution{}}
}

func (c *DistributionCache) get(relation rfqn.RelationFQN, now time.Time) (*distributions.Distribution, bool) {
	if now.Sub(c.loadedAt) > distributionCacheTTL {
		clear(c.entries)
		c.loadedAt = now
		return nil, false
	}
	ds, ok := c.entries[relation]
	return ds, ok
}

func (c *DistributionCache) put(relation rfqn.RelationFQN, ds *distributions.Distribution) {
	c.entries[relation] = ds
}
