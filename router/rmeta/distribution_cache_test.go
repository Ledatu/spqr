package rmeta

import (
	"context"
	"testing"
	"time"

	mockmgr "github.com/pg-sharding/spqr/pkg/mock/meta"
	"github.com/pg-sharding/spqr/pkg/models/distributions"
	"github.com/pg-sharding/spqr/pkg/session"
	"github.com/pg-sharding/spqr/router/rfqn"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type cachingSession struct {
	session.SessionParamsHolder
	cache *DistributionCache
}

func (s *cachingSession) RelationDistributionCache() *DistributionCache {
	return s.cache
}

func TestDistributionCacheExpires(t *testing.T) {
	cache := NewDistributionCache()
	relation := rfqn.RelationFQN{RelationName: "orders"}
	ds := &distributions.Distribution{Id: "ds1"}
	now := time.Now()

	_, ok := cache.get(relation, now)
	assert.False(t, ok)

	cache.put(relation, ds)
	got, ok := cache.get(relation, now.Add(distributionCacheTTL/2))
	assert.True(t, ok)
	assert.Same(t, ds, got)

	_, ok = cache.get(relation, now.Add(distributionCacheTTL+time.Second))
	assert.False(t, ok)
}

func TestGetRelationDistributionReusesClientCacheAcrossQueries(t *testing.T) {
	ctrl := gomock.NewController(t)
	mgr := mockmgr.NewMockEntityMgr(ctrl)
	relation := &rfqn.RelationFQN{RelationName: "orders"}
	ds := &distributions.Distribution{Id: "ds1"}
	mgr.EXPECT().GetRelationDistribution(gomock.Any(), relation).Return(ds, nil).Times(1)

	sph := &cachingSession{
		SessionParamsHolder: session.NewSimpleHandler("", false, "", ""),
		cache:               NewDistributionCache(),
	}
	for range 3 {
		rm := NewRoutingMetadataContext(sph, nil, "select 1", nil, nil, mgr)
		got, err := rm.GetRelationDistribution(context.Background(), relation)
		assert.NoError(t, err)
		assert.Same(t, ds, got)
	}
}

func TestGetRelationDistributionWithoutClientCacheQueriesEachContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	mgr := mockmgr.NewMockEntityMgr(ctrl)
	relation := &rfqn.RelationFQN{RelationName: "orders"}
	ds := &distributions.Distribution{Id: "ds1"}
	mgr.EXPECT().GetRelationDistribution(gomock.Any(), relation).Return(ds, nil).Times(2)

	sph := session.NewSimpleHandler("", false, "", "")
	for range 2 {
		rm := NewRoutingMetadataContext(sph, nil, "select 1", nil, nil, mgr)
		_, err := rm.GetRelationDistribution(context.Background(), relation)
		assert.NoError(t, err)
	}
}
