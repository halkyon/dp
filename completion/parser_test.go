package completion

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockDataProvider struct {
	commands       []string
	aliases        []string
	aliasesErr     error
	locations      []string
	locationsErr   error
	regions        []string
	regionsErr     error
	powerStatuses  []string
	serverStatuses []string
	names          []string
	namesErr       error
	tags           []string
	tagsErr        error
	fields         []string
}

func (m *mockDataProvider) GetCommands() []string { return m.commands }
func (m *mockDataProvider) GetAliases(ctx context.Context) ([]string, error) {
	return m.aliases, m.aliasesErr
}
func (m *mockDataProvider) GetLocations(ctx context.Context) ([]string, error) {
	return m.locations, m.locationsErr
}
func (m *mockDataProvider) GetRegions(ctx context.Context) ([]string, error) {
	return m.regions, m.regionsErr
}
func (m *mockDataProvider) GetPowerStatuses() []string  { return m.powerStatuses }
func (m *mockDataProvider) GetServerStatuses() []string { return m.serverStatuses }
func (m *mockDataProvider) GetNames(ctx context.Context) ([]string, error) {
	return m.names, m.namesErr
}
func (m *mockDataProvider) GetTags(ctx context.Context) ([]string, error) {
	return m.tags, m.tagsErr
}
func (m *mockDataProvider) GetFields() []string { return m.fields }

func defaultMock() *mockDataProvider {
	return &mockDataProvider{
		commands:       []string{"list", "create", "delete"},
		aliases:        []string{"web1", "web2", "db1"},
		locations:      []string{"ams1", "ams2", "lon1"},
		regions:        []string{"eu-west", "us-east"},
		powerStatuses:  []string{"on", "off", "rebooting"},
		serverStatuses: []string{"active", "pending", "terminated"},
		names:          []string{"server-a", "server-b"},
		tags:           []string{"prod", "staging", "dev"},
		fields:         []string{"id", "name", "status"},
	}
}

func TestNewParser(t *testing.T) {
	mock := defaultMock()
	filter := NewParser(mock)
	require.NotNil(t, filter)
}

func TestComplete_EmptyPrev_ReturnsCommands(t *testing.T) {
	mock := defaultMock()
	filter := NewParser(mock)

	result, err := filter.Complete(t.Context(), Request{Shell: ShellBash, Word: "", Prev: ""})
	require.NoError(t, err)
	assert.Equal(t, []string{"list", "create", "delete"}, result)
}

func TestComplete_FlagAliases(t *testing.T) {
	mock := defaultMock()
	filter := NewParser(mock)

	cases := []struct {
		name string
		prev string
	}{
		{"long flag", "--alias"},
		{"short flag", "-a"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := filter.Complete(t.Context(), Request{Shell: ShellBash, Word: "web", Prev: tc.prev})
			require.NoError(t, err)
			assert.Equal(t, []string{"web1", "web2"}, result)
		})
	}
}

func TestComplete_FlagLocations(t *testing.T) {
	mock := defaultMock()
	filter := NewParser(mock)

	cases := []struct {
		name string
		prev string
	}{
		{"long flag", "--location"},
		{"short flag", "-l"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := filter.Complete(t.Context(), Request{Shell: ShellBash, Word: "ams", Prev: tc.prev})
			require.NoError(t, err)
			assert.Equal(t, []string{"ams1", "ams2"}, result)
		})
	}
}

func TestComplete_FlagRegions(t *testing.T) {
	mock := defaultMock()
	filter := NewParser(mock)

	cases := []struct {
		name string
		prev string
	}{
		{"long flag", "--region"},
		{"short flag", "-r"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := filter.Complete(t.Context(), Request{Shell: ShellBash, Word: "eu", Prev: tc.prev})
			require.NoError(t, err)
			assert.Equal(t, []string{"eu-west"}, result)
		})
	}
}

func TestComplete_FlagPower(t *testing.T) {
	mock := defaultMock()
	filter := NewParser(mock)

	cases := []struct {
		name string
		prev string
	}{
		{"long flag", "--power"},
		{"short flag", "-p"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := filter.Complete(t.Context(), Request{Shell: ShellBash, Word: "on", Prev: tc.prev})
			require.NoError(t, err)
			assert.Equal(t, []string{"on"}, result)
		})
	}
}

func TestComplete_FlagStatus(t *testing.T) {
	mock := defaultMock()
	filter := NewParser(mock)

	cases := []struct {
		name string
		prev string
	}{
		{"long flag", "--status"},
		{"short flag", "-s"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := filter.Complete(t.Context(), Request{Shell: ShellBash, Word: "act", Prev: tc.prev})
			require.NoError(t, err)
			assert.Equal(t, []string{"active"}, result)
		})
	}
}

func TestComplete_FlagNames(t *testing.T) {
	mock := defaultMock()
	filter := NewParser(mock)

	cases := []struct {
		name string
		prev string
	}{
		{"long flag", "--name"},
		{"short flag", "-n"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := filter.Complete(t.Context(), Request{Shell: ShellBash, Word: "server", Prev: tc.prev})
			require.NoError(t, err)
			assert.Equal(t, []string{"server-a", "server-b"}, result)
		})
	}
}

func TestComplete_FlagTags(t *testing.T) {
	mock := defaultMock()
	filter := NewParser(mock)

	cases := []struct {
		name string
		prev string
	}{
		{"long flag", "--tag"},
		{"short flag", "-t"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := filter.Complete(t.Context(), Request{Shell: ShellBash, Word: "prod", Prev: tc.prev})
			require.NoError(t, err)
			assert.Equal(t, []string{"prod"}, result)
		})
	}
}

func TestComplete_FlagQuery(t *testing.T) {
	mock := defaultMock()
	filter := NewParser(mock)

	cases := []struct {
		name string
		prev string
	}{
		{"long flag", "--query"},
		{"short flag", "-q"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := filter.Complete(t.Context(), Request{Shell: ShellBash, Word: "id", Prev: tc.prev})
			require.NoError(t, err)
			assert.Equal(t, []string{"id"}, result)
		})
	}
}

func TestComplete_FlagSort(t *testing.T) {
	mock := defaultMock()
	filter := NewParser(mock)

	cases := []struct {
		name string
		prev string
	}{
		{"long flag", "--sort"},
		{"short flag", "-S"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := filter.Complete(t.Context(), Request{Shell: ShellBash, Word: "id", Prev: tc.prev})
			require.NoError(t, err)
			assert.Equal(t, []string{"id"}, result)
		})
	}
}

func TestComplete_FlagValueLongForm(t *testing.T) {
	mock := defaultMock()
	filter := NewParser(mock)

	cases := []struct {
		name     string
		prev     string
		word     string
		expected []string
	}{
		{"alias", "--alias=we", "", []string{"web1", "web2"}},
		{"location", "--location=am", "", []string{"ams1", "ams2"}},
		{"region", "--region=eu", "", []string{"eu-west"}},
		{"power", "--power=o", "", []string{"on", "off"}},
		{"status", "--status=act", "", []string{"active"}},
		{"name", "--name=server", "", []string{"server-a", "server-b"}},
		{"tag", "--tag=prod", "", []string{"prod"}},
		{"query", "--query=id", "", []string{"id"}},
		{"sort", "--sort=id", "", []string{"id"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := filter.Complete(t.Context(), Request{Shell: ShellBash, Word: tc.word, Prev: tc.prev})
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestComplete_FlagValueShortForm(t *testing.T) {
	mock := defaultMock()
	filter := NewParser(mock)

	cases := []struct {
		name     string
		prev     string
		word     string
		expected []string
	}{
		{"alias", "-awe", "", []string{"web1", "web2"}},
		{"location", "-lam", "", []string{"ams1", "ams2"}},
		{"region", "-reu", "", []string{"eu-west"}},
		{"power", "-po", "", []string{"on", "off"}},
		{"status", "-sact", "", []string{"active"}},
		{"name", "-nserver", "", []string{"server-a", "server-b"}},
		{"tag", "-tprod", "", []string{"prod"}},
		{"query", "-qid", "", []string{"id"}},
		{"sort", "-Sid", "", []string{"id"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := filter.Complete(t.Context(), Request{Shell: ShellBash, Word: tc.word, Prev: tc.prev})
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestComplete_SSHPrev_ReturnsAliases(t *testing.T) {
	mock := defaultMock()
	filter := NewParser(mock)

	result, err := filter.Complete(t.Context(), Request{Shell: ShellBash, Word: "web", Prev: "ssh"})
	require.NoError(t, err)
	assert.Equal(t, []string{"web1", "web2"}, result)
}

func TestComplete_DefaultFallback_ReturnsAliases(t *testing.T) {
	mock := defaultMock()
	filter := NewParser(mock)

	result, err := filter.Complete(t.Context(), Request{Shell: ShellBash, Word: "web", Prev: "something-random"})
	require.NoError(t, err)
	assert.Equal(t, []string{"web1", "web2"}, result)
}

func TestComplete_CaseInsensitiveFiltering(t *testing.T) {
	mock := defaultMock()
	filter := NewParser(mock)

	result, err := filter.Complete(t.Context(), Request{Shell: ShellBash, Word: "WEB", Prev: "--alias"})
	require.NoError(t, err)
	assert.Equal(t, []string{"web1", "web2"}, result)
}

func TestComplete_AliasAtPrefix(t *testing.T) {
	mock := defaultMock()
	filter := NewParser(mock)

	result, err := filter.Complete(t.Context(), Request{Shell: ShellBash, Word: "user@web", Prev: "--alias"})
	require.NoError(t, err)
	assert.Equal(t, []string{"web1", "web2"}, result)
}

func TestComplete_StripLeadingQuotes(t *testing.T) {
	mock := defaultMock()
	filter := NewParser(mock)

	cases := []struct {
		name string
		word string
	}{
		{"double quote", `"ams`},
		{"single quote", `'ams`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := filter.Complete(t.Context(), Request{Shell: ShellBash, Word: tc.word, Prev: "--location"})
			require.NoError(t, err)
			assert.Equal(t, []string{"ams1", "ams2"}, result)
		})
	}
}

func TestComplete_EmptyWord_ReturnsAllValues(t *testing.T) {
	mock := defaultMock()
	filter := NewParser(mock)

	result, err := filter.Complete(t.Context(), Request{Shell: ShellBash, Word: "", Prev: "--location"})
	require.NoError(t, err)
	assert.Equal(t, []string{"ams1", "ams2", "lon1"}, result)
}

func TestComplete_DataProviderError_ReturnsEmpty(t *testing.T) {
	mock := defaultMock()
	mock.aliasesErr = errors.New("network error")
	filter := NewParser(mock)

	result, err := filter.Complete(t.Context(), Request{Shell: ShellBash, Word: "web", Prev: "--alias"})
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestComplete_ZshFormat(t *testing.T) {
	mock := defaultMock()
	filter := NewParser(mock)

	result, err := filter.Complete(t.Context(), Request{Shell: ShellZsh, Word: "web", Prev: "--alias"})
	require.NoError(t, err)
	assert.Equal(t, []string{"web1", "web2"}, result)
}

func TestComplete_FishFormat(t *testing.T) {
	mock := defaultMock()
	filter := NewParser(mock)

	result, err := filter.Complete(t.Context(), Request{Shell: ShellFish, Word: "web", Prev: "--alias"})
	require.NoError(t, err)
	assert.Equal(t, []string{"web1", "web2"}, result)
}

func TestComplete_BashFormat(t *testing.T) {
	mock := defaultMock()
	filter := NewParser(mock)

	result, err := filter.Complete(t.Context(), Request{Shell: ShellBash, Word: "web", Prev: "--alias"})
	require.NoError(t, err)
	assert.Equal(t, []string{"web1", "web2"}, result)
}

func TestComplete_UnknownShell_DefaultsToBash(t *testing.T) {
	mock := defaultMock()
	filter := NewParser(mock)

	result, err := filter.Complete(t.Context(), Request{Shell: Shell("unknown"), Word: "web", Prev: "--alias"})
	require.NoError(t, err)
	assert.Equal(t, []string{"web1", "web2"}, result)
}

func TestComplete_FlagValueWithWord(t *testing.T) {
	mock := defaultMock()
	filter := NewParser(mock)

	// When prev contains partial value and word also has content, they should be combined
	result, err := filter.Complete(t.Context(), Request{Shell: ShellBash, Word: "1", Prev: "--alias=web"})
	require.NoError(t, err)
	assert.Equal(t, []string{"web1"}, result)
}

func TestComplete_CommandsCaseInsensitive(t *testing.T) {
	mock := defaultMock()
	filter := NewParser(mock)

	result, err := filter.Complete(t.Context(), Request{Shell: ShellBash, Word: "LI", Prev: ""})
	require.NoError(t, err)
	assert.Equal(t, []string{"list"}, result)
}

func TestComplete_EmptyWordCommands(t *testing.T) {
	mock := defaultMock()
	filter := NewParser(mock)

	result, err := filter.Complete(t.Context(), Request{Shell: ShellBash, Word: "", Prev: ""})
	require.NoError(t, err)
	assert.Equal(t, []string{"list", "create", "delete"}, result)
}

func TestComplete_PowerAndStatusNoContextRequired(t *testing.T) {
	mock := defaultMock()
	filter := NewParser(mock)

	// These don't require context, so they should work even if other methods error
	mock.aliasesErr = errors.New("fail")
	mock.locationsErr = errors.New("fail")

	result, err := filter.Complete(t.Context(), Request{Shell: ShellBash, Word: "", Prev: "--power"})
	require.NoError(t, err)
	assert.Equal(t, []string{"on", "off", "rebooting"}, result)

	result, err = filter.Complete(t.Context(), Request{Shell: ShellBash, Word: "", Prev: "--status"})
	require.NoError(t, err)
	assert.Equal(t, []string{"active", "pending", "terminated"}, result)
}
