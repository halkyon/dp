package shell

import (
	"errors"
	"testing"

	"github.com/halkyon/dp/completion"
	"github.com/stretchr/testify/require"
)

func TestDetectShell_FromEnv(t *testing.T) {
	tests := []struct {
		name    string
		env     string
		want    completion.Shell
		wantErr bool
	}{
		{"bash from env", "/bin/bash", completion.ShellBash, false},
		{"zsh from env", "/usr/bin/zsh", completion.ShellZsh, false},
		{"fish from env", "/usr/local/bin/fish", completion.ShellFish, false},
		{"unsupported env", "/bin/csh", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SHELL", tt.env)

			oldDetect := detectParentShell
			detectParentShell = func() (string, error) { return "", nil }
			defer func() { detectParentShell = oldDetect }()

			got, err := DetectShell()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestDetectShell_FromParent(t *testing.T) {
	tests := []struct {
		name    string
		parent  string
		want    completion.Shell
		wantErr bool
	}{
		{"bash from parent", "bash", completion.ShellBash, false},
		{"zsh from parent", "zsh", completion.ShellZsh, false},
		{"fish from parent", "fish", completion.ShellFish, false},
		{"unsupported parent", "csh", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SHELL", "")

			oldDetect := detectParentShell
			detectParentShell = func() (string, error) { return tt.parent, nil }
			defer func() { detectParentShell = oldDetect }()

			got, err := DetectShell()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestDetectShell_ParentDetectionFails(t *testing.T) {
	t.Setenv("SHELL", "")

	oldDetect := detectParentShell
	detectParentShell = func() (string, error) { return "", errors.New("no parent") }
	defer func() { detectParentShell = oldDetect }()

	_, err := DetectShell()
	require.Error(t, err)
}
