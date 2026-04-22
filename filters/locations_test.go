package filters

import (
	"context"
	"testing"
	"time"

	"github.com/halkyon/dp/testapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocations_Get(t *testing.T) {
	var mq testapi.MockQuerier

	locs, err := NewLocations(&mq, time.Hour, t.TempDir())
	require.NoError(t, err)
	require.NoError(t, locs.Clear())

	t.Run("First call (cache miss)", func(t *testing.T) {
		locations, err := locs.Get(context.Background())
		require.NoError(t, err)
		assert.Equal(t, []string{"Amsterdam", "Ashburn", "Atlanta", "Berlin", "Chicago", "Dallas", "Frankfurt", "Hong Kong", "London", "Los Angeles", "Miami", "New York", "Paris", "Seattle", "Singapore", "Sydney", "Tokyo", "Toronto"}, locations)
	})

	t.Run("Second call (cache hit)", func(t *testing.T) {
		locations, err := locs.Get(context.Background())
		require.NoError(t, err)
		assert.Equal(t, []string{"Amsterdam", "Ashburn", "Atlanta", "Berlin", "Chicago", "Dallas", "Frankfurt", "Hong Kong", "London", "Los Angeles", "Miami", "New York", "Paris", "Seattle", "Singapore", "Sydney", "Tokyo", "Toronto"}, locations)
	})
}

func TestRegions_Get(t *testing.T) {
	var mq testapi.MockQuerier

	regions, err := NewRegions(&mq, time.Hour, t.TempDir())
	require.NoError(t, err)

	t.Run("First call (cache miss)", func(t *testing.T) {
		regionList, err := regions.Get(context.Background())
		require.NoError(t, err)
		assert.Equal(t, []string{"AP", "EU", "NA"}, regionList)
	})

	t.Run("Second call (cache hit)", func(t *testing.T) {
		regionList, err := regions.Get(context.Background())
		require.NoError(t, err)
		assert.Equal(t, []string{"AP", "EU", "NA"}, regionList)
	})
}
