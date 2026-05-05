package shell

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/halkyon/dp/completion"
)

var detectParentShell = func() (string, error) {
	ppid := os.Getppid()
	switch runtime.GOOS {
	case "linux":
		data, err := os.ReadFile("/proc/" + strconv.Itoa(ppid) + "/comm")
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(data)), nil
	case "darwin":
		out, err := exec.Command("ps", "-p", strconv.Itoa(ppid), "-o", "comm=").Output()
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(out)), nil
	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func mapShell(name string) (completion.Shell, error) {
	switch strings.ToLower(name) {
	case "bash":
		return completion.ShellBash, nil
	case "zsh":
		return completion.ShellZsh, nil
	case "fish":
		return completion.ShellFish, nil
	default:
		return "", fmt.Errorf("unsupported shell: %s", name)
	}
}

func DetectShell() (completion.Shell, error) {
	if shellEnv := os.Getenv("SHELL"); shellEnv != "" {
		return mapShell(filepath.Base(shellEnv))
	}

	parentShell, err := detectParentShell()
	if err != nil {
		return "", fmt.Errorf("failed to detect parent shell: %w", err)
	}

	return mapShell(filepath.Base(parentShell))
}
