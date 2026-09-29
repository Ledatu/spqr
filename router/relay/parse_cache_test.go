package relay

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCacheParseKeepsEntryCountBounded(t *testing.T) {
	rst := &RelayStateImpl{parseCache: map[string]ParseCacheEntry{}}

	for i := 0; i < 10*parseCacheMaxEntries; i++ {
		query := fmt.Sprintf("select * from t where id in (%d)", i)
		rst.cacheParse(query, ParseCacheEntry{})

		assert.LessOrEqual(t, len(rst.parseCache), parseCacheMaxEntries)
		assert.Contains(t, rst.parseCache, query)
	}
}

func TestCacheParseKeepsQueryBytesBounded(t *testing.T) {
	rst := &RelayStateImpl{parseCache: map[string]ParseCacheEntry{}}
	half := strings.Repeat("x", parseCacheMaxBytes/2+1)

	rst.cacheParse(half+"a", ParseCacheEntry{})
	rst.cacheParse(half+"b", ParseCacheEntry{})

	assert.Len(t, rst.parseCache, 1)
	assert.Contains(t, rst.parseCache, half+"b")
	assert.Equal(t, len(half)+1, rst.parseCacheBytes)
}

func TestCacheParseKeepsOversizedQueryUntilNextParse(t *testing.T) {
	rst := &RelayStateImpl{parseCache: map[string]ParseCacheEntry{}}
	huge := strings.Repeat("x", 2*parseCacheMaxBytes)

	rst.cacheParse("select 1", ParseCacheEntry{})
	rst.cacheParse(huge, ParseCacheEntry{})
	assert.Len(t, rst.parseCache, 1)
	assert.Contains(t, rst.parseCache, huge)

	rst.cacheParse("select 2", ParseCacheEntry{})
	assert.Len(t, rst.parseCache, 1)
	assert.Contains(t, rst.parseCache, "select 2")
}
