package server

import (
	"testing"

	"github.com/halkyon/dp/testapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildQuery(t *testing.T) {
	t.Run("full query when no fields specified", func(t *testing.T) {
		q := buildQuery(nil)
		assert.Contains(t, q, "name")
		assert.Contains(t, q, "alias")
		assert.Contains(t, q, "location {")
		assert.Contains(t, q, "hardware {")
		assert.Contains(t, q, "billing {")
	})

	t.Run("minimal query for basic fields", func(t *testing.T) {
		q := buildQuery([]string{"Name", "Alias", "Status"})
		assert.Contains(t, q, "name")
		assert.Contains(t, q, "alias")
		assert.Contains(t, q, "statusV2")
		assert.NotContains(t, q, "location {")
		assert.NotContains(t, q, "hardware {")
		assert.NotContains(t, q, "billing {")
	})

	t.Run("query includes nested blocks for requested fields", func(t *testing.T) {
		q := buildQuery([]string{"Name", "Location", "CPU", "Price"})
		assert.Contains(t, q, "name")
		assert.Contains(t, q, "location {")
		assert.Contains(t, q, "hardware {")
		assert.Contains(t, q, "billing {")
		assert.NotContains(t, q, "system {")
		assert.NotContains(t, q, "network {")
	})

	t.Run("IP field includes network block", func(t *testing.T) {
		q := buildQuery([]string{"Name", "IP"})
		assert.Contains(t, q, "network {")
		assert.Contains(t, q, "ipAddresses")
	})

	t.Run("Tags field includes tags block", func(t *testing.T) {
		q := buildQuery([]string{"Name", "Tags"})
		assert.Contains(t, q, "tags {")
		assert.NotContains(t, q, "hardware {")
	})
}

func TestServer_List(t *testing.T) {
	var mq testapi.MockQuerier

	t.Run("List servers", func(t *testing.T) {
		servers, err := List(t.Context(), &mq)
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
		servers, err := List(t.Context(), &mq, WithLocation("Amsterdam"), WithPower("ON"))
		require.NoError(t, err)
		assert.Len(t, servers, 1)
		assert.Equal(t, "DP-12345", servers[0].Name)
	})

	t.Run("Filter by power OFF", func(t *testing.T) {
		servers, err := List(t.Context(), &mq, WithPower("OFF"))
		require.NoError(t, err)
		assert.Len(t, servers, 1)
		assert.Equal(t, "DP-11111", servers[0].Name)
	})

	t.Run("Filter by region", func(t *testing.T) {
		servers, err := List(t.Context(), &mq, WithRegion("EU"))
		require.NoError(t, err)
		assert.Len(t, servers, 1)
		assert.Equal(t, "DP-12345", servers[0].Name)
	})

	t.Run("Filter by status", func(t *testing.T) {
		servers, err := List(t.Context(), &mq, WithStatus("PROVISIONING"))
		require.NoError(t, err)
		assert.Len(t, servers, 1)
		assert.Equal(t, "DP-11111", servers[0].Name)
	})

	t.Run("Filter by name", func(t *testing.T) {
		servers, err := List(t.Context(), &mq, WithName("DP-67890"))
		require.NoError(t, err)
		assert.Len(t, servers, 1)
		assert.Equal(t, "DP-67890", servers[0].Name)
	})

	t.Run("Filter by alias", func(t *testing.T) {
		servers, err := List(t.Context(), &mq, WithAlias("test-server-1"))
		require.NoError(t, err)
		assert.Len(t, servers, 1)
		assert.Equal(t, "DP-12345", servers[0].Name)
	})

	t.Run("Filter by region and power", func(t *testing.T) {
		servers, err := List(t.Context(), &mq, WithRegion("NA"), WithPower("ON"))
		require.NoError(t, err)
		assert.Len(t, servers, 1)
		assert.Equal(t, "DP-67890", servers[0].Name)
	})

	t.Run("WithFields requests only specified data", func(t *testing.T) {
		servers, err := List(t.Context(), &mq, WithFields("Name", "Alias", "Status"))
		require.NoError(t, err)
		assert.Len(t, servers, 3)

		for _, s := range servers {
			assert.NotEmpty(t, s.Name)
			assert.NotEmpty(t, s.Status)
			// These fields were not requested so the mock won't return them
			assert.Empty(t, s.IP)
			assert.Empty(t, s.Location)
			assert.Empty(t, s.CPU)
			assert.Empty(t, s.Memory)
			assert.Zero(t, s.Price)
		}
	})

	t.Run("WithFields includes nested blocks correctly", func(t *testing.T) {
		servers, err := List(t.Context(), &mq, WithFields("Name", "IP", "Location"))
		require.NoError(t, err)
		require.Len(t, servers, 3)

		assert.NotEmpty(t, servers[0].Name)
		assert.NotEmpty(t, servers[0].IP)
		assert.NotEmpty(t, servers[0].Location)
		assert.Empty(t, servers[0].CPU)
		assert.Empty(t, servers[0].OperatingSystem)
	})

	t.Run("WithFields combined with filters", func(t *testing.T) {
		servers, err := List(t.Context(), &mq,
			WithLocation("Amsterdam"),
			WithFields("Name", "Alias", "Location"),
		)
		require.NoError(t, err)
		assert.Len(t, servers, 1)
		assert.Equal(t, "DP-12345", servers[0].Name)
		assert.Equal(t, "test-server-1", servers[0].Alias)
		assert.Equal(t, "Amsterdam", servers[0].Location)
		assert.Empty(t, servers[0].IP)
	})
}
