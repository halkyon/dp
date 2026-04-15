package server

import (
	"context"
	"fmt"
	"testing"

	"github.com/halkyon/dp/api"
	"github.com/halkyon/dp/testapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer_List(t *testing.T) {
	srv, err := testapi.Start(t.Context())
	require.NoError(t, err)

	url := fmt.Sprintf("http://%s", srv.Addr())

	client, err := api.NewClient("test-key")
	require.NoError(t, err)
	client.SetBaseURL(url)

	t.Run("List servers", func(t *testing.T) {
		servers, err := List(context.Background(), client)
		assert.NoError(t, err)
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
		assert.Equal(t, 49.99, server1.Price)

		assert.Equal(t, "test-server-2", server2.Alias)
		assert.Equal(t, "192.168.2.1", server2.IP)
		assert.Equal(t, "Debian 11", server2.OperatingSystem)
		assert.Equal(t, "AMD EPYC 7443", server2.CPU)
		assert.Equal(t, "64 GB", server2.Memory)
		assert.Equal(t, 149.99, server2.Price)

		assert.Equal(t, "", server3.Alias)
		assert.Equal(t, "2001:db8::1", server3.IP)
		assert.Equal(t, "CentOS 8", server3.OperatingSystem)
		assert.Equal(t, "Intel Xeon Gold 6330", server3.CPU)
		assert.Equal(t, "128 GB", server3.Memory)
		assert.Equal(t, "960 GB NVMe", server3.Storage)
		assert.Equal(t, 299.99, server3.Price)
	})
}
