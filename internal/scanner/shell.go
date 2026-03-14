package scanner

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/tool-chain-brain/tcb/pkg/models"
)

// envVarsToCapture lists environment variables relevant to a developer toolchain.
var envVarsToCapture = []string{
	"GOPATH", "GOROOT", "GOBIN",
	"CARGO_HOME", "RUSTUP_HOME",
	"PYENV_ROOT", "VIRTUAL_ENV",
	"NVM_DIR", "NODE_PATH",
	"JAVA_HOME", "ANDROID_HOME", "ANDROID_SDK_ROOT",
	"FLUTTER_ROOT", "DART_HOME",
	"SDKMAN_DIR",
	"EDITOR", "VISUAL",
	"DOCKER_HOST", "DOCKER_BUILDKIT",
	"KUBECONFIG",
	"AWS_DEFAULT_REGION", "AWS_PROFILE",
	"GCLOUD_PROJECT",
	"CI", "GITHUB_ACTIONS",
}

// ScanShell captures the active shell, config files, and key environment variables.
func ScanShell() models.ShellConfig {
	cfg := models.ShellConfig{}

	shellPath := os.Getenv("SHELL")
	if shellPath == "" {
		shellPath = detectShellFromPath()
	}
	cfg.Shell = shellPath

	// Get shell version
	if shellPath != "" {
		shellBin := filepath.Base(shellPath)
		switch shellBin {
		case "bash":
			_, raw := runCmd("bash", "--version")
			// "GNU bash, version 5.2.15(1)-release ..."
			if idx := strings.Index(raw, "version "); idx != -1 {
				rest := raw[idx+8:]
				parts := strings.Fields(rest)
				if len(parts) > 0 {
					cfg.Version = strings.Split(parts[0], "(")[0]
				}
			}
		case "zsh":
			_, raw := runCmd("zsh", "--version")
			// "zsh 5.9 ..."
			parts := strings.Fields(raw)
			if len(parts) >= 2 {
				cfg.Version = parts[1]
			}
		case "fish":
			_, raw := runCmd("fish", "--version")
			cfg.Version = strings.TrimPrefix(firstLine(raw), "fish, version ")
		case "sh", "dash":
			_, raw := runCmd(shellBin, "--version")
			cfg.Version = firstLine(raw)
		}
	}

	// Collect config files
	home, _ := os.UserHomeDir()
	candidates := []string{
		filepath.Join(home, ".bashrc"),
		filepath.Join(home, ".bash_profile"),
		filepath.Join(home, ".bash_aliases"),
		filepath.Join(home, ".zshrc"),
		filepath.Join(home, ".zprofile"),
		filepath.Join(home, ".zshenv"),
		filepath.Join(home, ".profile"),
		filepath.Join(home, ".config", "fish", "config.fish"),
		filepath.Join(home, ".config", "fish", "conf.d"),
		filepath.Join(home, ".tcshrc"),
		filepath.Join(home, ".cshrc"),
		filepath.Join(home, ".kshrc"),
	}
	for _, f := range candidates {
		if _, err := os.Stat(f); err == nil {
			cfg.ConfigFiles = append(cfg.ConfigFiles, f)
		}
	}

	// Capture selected env vars
	cfg.EnvVars = make(map[string]string)
	for _, key := range envVarsToCapture {
		if val := os.Getenv(key); val != "" {
			cfg.EnvVars[key] = val
		}
	}

	// Capture PATH entries
	pathRaw := os.Getenv("PATH")
	if pathRaw != "" {
		entries := strings.Split(pathRaw, string(os.PathListSeparator))
		cfg.PathEntries = entries
	}

	return cfg
}

func detectShellFromPath() string {
	for _, candidate := range []string{"zsh", "bash", "fish", "sh"} {
		if p, _ := runCmd(candidate, "--version"); p != "" {
			return p
		}
	}
	return ""
}
