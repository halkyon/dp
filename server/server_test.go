package server

import (
	"context"
	"testing"

	"github.com/halkyon/dp/testapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer_List(t *testing.T) {
	var mq testapi.MockQuerier

	t.Run("List servers", func(t *testing.T) {
		servers, err := List(context.Background(), &mq)
		require.NoError(t, err)
		assert.Len(t, servers, 3)

		// Find server by name
		var server1, server2, server3 Server
		for _, s := range servers {
			switch s.Name {
			case "DP-12345":
				server1 = s
			case "DP-67890":
				server2 = s
			case "DP-11111":
				server3 = s
			}
		}

		assert.Equal(t, "test-server-1", server1.Alias)
		assert.Equal(t, "192.168.1.1", server1.IP)
		assert.Equal(t, "Ubuntu 22.04", server1.OperatingSystem)
		assert.Equal(t, "Intel Xeon E-2388", server1.CPU)
		assert.Equal(t, "32 GB", server1.Memory)
		assert.Equal(t, "512 GB NVMe", server1.Storage)
		assert.InDelta(t, 49.99, server1.Price, 0.01)

		assert.Equal(t, "test-server-2", server2.Alias)
		assert.Equal(t, "192.168.2.1", server2.IP)
		assert.Equal(t, "Debian 11", server2.OperatingSystem)
		assert.Equal(t, "AMD EPYC 7443", server2.CPU)
		assert.Equal(t, "64 GB", server2.Memory)
		assert.InDelta(t, 149.99, server2.Price, 0.01)

		assert.Empty(t, server3.Alias)
		assert.Equal(t, "2001:db8::1", server3.IP)
		assert.Equal(t, "CentOS 8", server3.OperatingSystem)
		assert.Equal(t, "Intel Xeon Gold 6330", server3.CPU)
		assert.Equal(t, "128 GB", server3.Memory)
		assert.Equal(t, "960 GB NVMe", server3.Storage)
		assert.InDelta(t, 299.99, server3.Price, 0.01)
	})

	t.Run("Filter by location and power", func(t *testing.T) {
		servers, err := List(context.Background(), &mq, WithLocation("Amsterdam"), WithPower("ON"))
		require.NoError(t, err)
		assert.Len(t, servers, 1)
		assert.Equal(t, "DP-12345", servers[0].Name)
	})

	t.Run("Filter by power OFF", func(t *testing.T) {
		servers, err := List(context.Background(), &mq, WithPower("OFF"))
		require.NoError(t, err)
		assert.Len(t, servers, 1)
		assert.Equal(t, "DP-11111", servers[0].Name)
	})

	t.Run("Filter by region", func(t *testing.T) {
		servers, err := List(context.Background(), &mq, WithRegion("EU"))
		require.NoError(t, err)
		assert.Len(t, servers, 1)
		assert.Equal(t, "DP-12345", servers[0].Name)
	})

	t.Run("Filter by status", func(t *testing.T) {
		servers, err := List(context.Background(), &mq, WithStatus("PROVISIONING"))
		require.NoError(t, err)
		assert.Len(t, servers, 1)
		assert.Equal(t, "DP-11111", servers[0].Name)
	})

	t.Run("Filter by name", func(t *testing.T) {
		servers, err := List(context.Background(), &mq, WithName("DP-67890"))
		require.NoError(t, err)
		assert.Len(t, servers, 1)
		assert.Equal(t, "DP-67890", servers[0].Name)
	})

	t.Run("Filter by alias", func(t *testing.T) {
		servers, err := List(context.Background(), &mq, WithAlias("test-server-1"))
		require.NoError(t, err)
		assert.Len(t, servers, 1)
		assert.Equal(t, "DP-12345", servers[0].Name)
	})

	t.Run("Filter by region and power", func(t *testing.T) {
		servers, err := List(context.Background(), &mq, WithRegion("NA"), WithPower("ON"))
		require.NoError(t, err)
		assert.Len(t, servers, 1)
		assert.Equal(t, "DP-67890", servers[0].Name)
	})
}
