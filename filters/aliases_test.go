package filters

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/halkyon/dp/api"
	"github.com/halkyon/dp/testapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAliases_Get(t *testing.T) {
	srv, err := testapi.Start(t.Context())
	require.NoError(t, err)

	url := fmt.Sprintf("http://%s", srv.Addr())

	client, err := api.NewClient("test-key")
	require.NoError(t, err)
	client.SetBaseURL(url)

	cache := NewAliases(client, time.Hour)

	t.Run("First call (cache miss)", func(t *testing.T) {
		aliases, err := cache.Get(context.Background())
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"test-server-1", "test-server-2"}, aliases)
	})

	t.Run("Second call (cache hit)", func(t *testing.T) {
		aliases, err := cache.Get(context.Background())
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"test-server-1", "test-server-2"}, aliases)
	})
}
