package filters

import (
	"testing"

	"github.com/halkyon/dp/testapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNames_Get(t *testing.T) {
	var mq testapi.MockQuerier

	names := NewNames(&mq, 0)

	result, err := names.Get(t.Context())
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"DP-12345", "DP-67890", "DP-11111"}, result)
}
