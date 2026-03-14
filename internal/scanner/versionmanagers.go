package scanner

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/tool-chain-brain/tcb/pkg/models"
)

// ScanVersionManagers detects runtime version managers (nvm, pyenv, asdf, etc.)
func ScanVersionManagers() []models.Tool {
	home, _ := os.UserHomeDir()

	var tools []models.Tool

	// nvm — not on PATH by default, check the directory
	nvmDir := os.Getenv("NVM_DIR")
	if nvmDir == "" {
		nvmDir = filepath.Join(home, ".nvm")
	}
	if stat, err := os.Stat(nvmDir); err == nil && stat.IsDir() {
		version := ""
		if path, raw := runCmd("nvm", "--version"); path != "" {
			version = firstLine(raw)
		} else {
			// Try reading the alias/default version file
			defaultFile := filepath.Join(nvmDir, "alias", "default")
			if b, err := os.ReadFile(defaultFile); err == nil {
				version = strings.TrimSpace(string(b))
			}
		}
		tools = append(tools, models.Tool{
			Name:    "nvm",
			Version: version,
			Path:    nvmDir,
			Extra:   map[string]string{"nvm_dir": nvmDir},
		})
	}

	// volta
	if path, raw := runCmd("volta", "--version"); path != "" {
		tools = append(tools, models.Tool{
			Name:    "volta",
			Version: firstLine(raw),
			Path:    path,
		})
	}

	// fnm
	if path, raw := runCmd("fnm", "--version"); path != "" {
		tools = append(tools, models.Tool{
			Name:    "fnm",
			Version: strings.TrimPrefix(firstLine(raw), "fnm "),
			Path:    path,
		})
	}

	// pyenv
	pyenvRoot := os.Getenv("PYENV_ROOT")
	if pyenvRoot == "" {
		pyenvRoot = filepath.Join(home, ".pyenv")
	}
	if path, raw := runCmd("pyenv", "--version"); path != "" {
		tools = append(tools, models.Tool{
			Name:    "pyenv",
			Version: strings.TrimPrefix(firstLine(raw), "pyenv "),
			Path:    path,
			Extra:   map[string]string{"pyenv_root": pyenvRoot},
		})
	} else if stat, err := os.Stat(pyenvRoot); err == nil && stat.IsDir() {
		tools = append(tools, models.Tool{
			Name:  "pyenv",
			Path:  pyenvRoot,
			Extra: map[string]string{"pyenv_root": pyenvRoot},
		})
	}

	// rbenv
	if path, raw := runCmd("rbenv", "--version"); path != "" {
		tools = append(tools, models.Tool{
			Name:    "rbenv",
			Version: strings.TrimPrefix(firstLine(raw), "rbenv "),
			Path:    path,
		})
	}

	// rvm
	if path, raw := runCmd("rvm", "--version"); path != "" {
		tools = append(tools, models.Tool{
			Name:    "rvm",
			Version: firstLine(raw),
			Path:    path,
		})
	}

	// asdf
	if path, raw := runCmd("asdf", "version"); path != "" {
		tools = append(tools, models.Tool{
			Name:    "asdf",
			Version: strings.TrimPrefix(firstLine(raw), "v"),
			Path:    path,
		})
	}

	// mise (formerly rtx)
	if path, raw := runCmd("mise", "--version"); path != "" {
		tools = append(tools, models.Tool{
			Name:    "mise",
			Version: firstLine(raw),
			Path:    path,
		})
	}

	// sdkman
	sdkmanDir := filepath.Join(home, ".sdkman")
	if stat, err := os.Stat(sdkmanDir); err == nil && stat.IsDir() {
		version := ""
		versionFile := filepath.Join(sdkmanDir, "var", "version")
		if b, err := os.ReadFile(versionFile); err == nil {
			version = strings.TrimSpace(string(b))
		}
		tools = append(tools, models.Tool{
			Name:    "sdkman",
			Version: version,
			Path:    sdkmanDir,
		})
	}

	// rustup
	if path, raw := runCmd("rustup", "--version"); path != "" {
		tools = append(tools, models.Tool{
			Name:    "rustup",
			Version: strings.TrimPrefix(firstLine(raw), "rustup "),
			Path:    path,
		})
	}

	// goenv
	if path, raw := runCmd("goenv", "--version"); path != "" {
		tools = append(tools, models.Tool{
			Name:    "goenv",
			Version: strings.TrimPrefix(firstLine(raw), "goenv "),
			Path:    path,
		})
	}

	// jenv
	if path, raw := runCmd("jenv", "--version"); path != "" {
		tools = append(tools, models.Tool{
			Name:    "jenv",
			Version: firstLine(raw),
			Path:    path,
		})
	}

	return tools
}
