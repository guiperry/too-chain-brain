package export

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tool-chain-brain/tcb/pkg/models"
)

// devcontainerJSON is the structure for .devcontainer/devcontainer.json
type devcontainerJSON struct {
	Name           string             `json:"name"`
	Build          devcontainerBuild  `json:"build"`
	Features       map[string]any     `json:"features,omitempty"`
	ForwardPorts   []int              `json:"forwardPorts,omitempty"`
	PostCreateCmd  string             `json:"postCreateCommand,omitempty"`
	RemoteUser     string             `json:"remoteUser"`
	Customizations devcontainerCustom `json:"customizations,omitempty"`
	Mounts         []string           `json:"mounts,omitempty"`
	ContainerEnv   map[string]string  `json:"containerEnv,omitempty"`
}

type devcontainerBuild struct {
	Dockerfile string            `json:"dockerfile"`
	Args       map[string]string `json:"args,omitempty"`
}

type devcontainerCustom struct {
	VSCode devcontainerVSCode `json:"vscode,omitempty"`
}

type devcontainerVSCode struct {
	Extensions []string       `json:"extensions,omitempty"`
	Settings   map[string]any `json:"settings,omitempty"`
}

// WriteDevContainer generates .devcontainer/devcontainer.json in outDir.
func WriteDevContainer(tc *models.Toolchain, outDir string) (string, error) {
	devDir := filepath.Join(outDir, ".devcontainer")
	if err := os.MkdirAll(devDir, 0755); err != nil {
		return "", err
	}

	// Build features map for any tools not covered by the Dockerfile
	features := make(map[string]any)

	for _, vm := range tc.VersionManagers {
		switch vm.Name {
		case "nvm":
			features["ghcr.io/devcontainers/features/node:1"] = map[string]string{
				"version": "lts",
			}
		}
	}

	// Collect extensions from VS Code scan
	var extensions []string
	if tc.Editor.VSCode != nil {
		extensions = tc.Editor.VSCode.Extensions
	}

	// Sensible default VS Code settings
	vscSettings := map[string]any{
		"terminal.integrated.defaultProfile.linux": "bash",
		"editor.formatOnSave":                      true,
		"editor.tabSize":                           2,
	}

	// Adjust tab size based on detected languages
	for _, lang := range tc.Languages {
		if lang.Name == "go" {
			vscSettings["editor.tabSize"] = 4
			vscSettings["[go]"] = map[string]any{
				"editor.insertSpaces": false,
			}
		}
		if lang.Name == "python3" || lang.Name == "python" {
			vscSettings["[python]"] = map[string]any{
				"editor.tabSize": 4,
			}
		}
	}

	// Build container env from important captured env vars
	containerEnv := make(map[string]string)
	importantEnvKeys := []string{"GOPATH", "GOBIN", "CARGO_HOME", "PYENV_ROOT", "NVM_DIR"}
	for _, k := range importantEnvKeys {
		if v, ok := tc.Shell.EnvVars[k]; ok && v != "" {
			containerEnv[k] = v
		}
	}
	if len(containerEnv) == 0 {
		containerEnv = nil
	}

	if len(features) == 0 {
		features = nil
	}

	dc := devcontainerJSON{
		Name: fmt.Sprintf("%s — TCB Dev Environment", tc.Meta.Hostname),
		Build: devcontainerBuild{
			Dockerfile: "../Dockerfile.devenv",
			Args: map[string]string{
				"USERNAME": "dev",
			},
		},
		Features:      features,
		RemoteUser:    "dev",
		PostCreateCmd: "echo '✅ Tool-Chain-Brain devcontainer ready'",
		Customizations: devcontainerCustom{
			VSCode: devcontainerVSCode{
				Extensions: extensions,
				Settings:   vscSettings,
			},
		},
		ContainerEnv: containerEnv,
	}

	dest := filepath.Join(devDir, "devcontainer.json")
	f, err := os.Create(dest)
	if err != nil {
		return "", err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(dc); err != nil {
		return "", err
	}

	return dest, nil
}
