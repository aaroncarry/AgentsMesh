package envpath

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// pathSentinel is a unique prefix printed before $PATH so we can extract it
// even when the login shell profile emits other output (nvm init, banners, etc.).
const pathSentinel = "AGENTSMESH_PATH="

// ResolveLoginShellPATH resolves the user's login shell PATH by spawning
// a login shell. This is critical when the runner runs as a systemd/launchd
// service, which provides only a minimal PATH (e.g. /usr/bin:/bin).
//
// On any failure, it falls back to the current process PATH.
func ResolveLoginShellPATH() string {
	fallback := os.Getenv("PATH")

	shell := os.Getenv("SHELL")
	if shell == "" {
		slog.Warn("envpath: $SHELL not set, using current PATH")
		return fallback
	}

	if _, err := exec.LookPath(shell); err != nil {
		slog.Warn("envpath: shell binary not found, using current PATH", "shell", shell, "error", err)
		return fallback
	}

	// Fish shell uses space-separated PATH; skip login resolution and use current PATH.
	if filepath.Base(shell) == "fish" {
		slog.Info("envpath: fish shell detected, using current PATH")
		return fallback
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use a sentinel-prefixed printf so that noisy profile scripts (nvm, welcome
	// banners, etc.) don't corrupt the resolved value. We scan stdout line-by-line
	// and extract only the line that starts with the sentinel.
	cmd := exec.CommandContext(ctx, shell, "-l", "-c",
		fmt.Sprintf("printf '%s%%s\\n' \"$PATH\"", pathSentinel))
	cmd.Env = []string{
		"HOME=" + os.Getenv("HOME"),
		"USER=" + os.Getenv("USER"),
		"LOGNAME=" + os.Getenv("LOGNAME"),
		"SHELL=" + shell,
		"TERM=dumb",
	}

	out, err := cmd.Output()
	if err != nil {
		slog.Warn("envpath: failed to resolve login shell PATH, using current PATH", "shell", shell, "error", err)
		return fallback
	}

	var resolved string
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, pathSentinel) {
			resolved = strings.TrimPrefix(line, pathSentinel)
			break
		}
	}

	// Validate: a real PATH must be non-empty.
	if resolved == "" {
		slog.Warn("envpath: login shell returned empty PATH, using current PATH")
		return fallback
	}

	dirs := strings.Split(resolved, ":")
	slog.Info("envpath: resolved login shell PATH", "dirs", len(dirs))

	return resolved
}
