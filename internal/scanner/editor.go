package scanner

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tool-chain-brain/tcb/pkg/models"
)

// ScanEditor detects VS Code and any .editorconfig files.
func ScanEditor() models.EditorConfig {
	cfg := models.EditorConfig{}

	// VS Code — check both `code` and `code-insiders`
	for _, bin := range []string{"code", "code-insiders"} {
		if _, err := exec.LookPath(bin); err == nil {
			vsc := &models.VSCodeConfig{}

			// Version
			if _, raw := runCmd(bin, "--version"); raw != "" {
				vsc.Version = firstLine(raw)
			}

			// Installed extensions
			if out, err := exec.Command(bin, "--list-extensions", "--show-versions").Output(); err == nil {
				for _, line := range strings.Split(string(out), "\n") {
					line = strings.TrimSpace(line)
					if line != "" {
						vsc.Extensions = append(vsc.Extensions, line)
					}
				}
			}

			cfg.VSCode = vsc
			break
		}
	}

	// .editorconfig — check home dir and CWD
	home, _ := os.UserHomeDir()
	cwd, _ := os.Getwd()

	for _, dir := range []string{home, cwd} {
		if dir == "" {
			continue
		}
		candidate := filepath.Join(dir, ".editorconfig")
		if _, err := os.Stat(candidate); err == nil {
			cfg.EditorConfigFile = candidate
			break
		}
	}

	return cfg
}
