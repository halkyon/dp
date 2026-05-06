package filters

import (
	"testing"

	"github.com/halkyon/dp/testapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTags_Get(t *testing.T) {
	var mq testapi.MockQuerier

	tags := NewTags(&mq, 0)

	result, err := tags.Get(t.Context())
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"env=production"}, result)
}
