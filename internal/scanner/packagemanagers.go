package scanner

import (
	"strings"

	"github.com/tool-chain-brain/tcb/pkg/models"
)

// ScanPackageManagers probes all known package managers.
func ScanPackageManagers() []models.Tool {
	type pmDef struct {
		name    string
		cmd     string
		args    []string
		extract func(string) string
	}

	defs := []pmDef{
		{
			name:    "npm",
			cmd:     "npm",
			args:    []string{"--version"},
			extract: firstLine,
		},
		{
			name:    "yarn",
			cmd:     "yarn",
			args:    []string{"--version"},
			extract: firstLine,
		},
		{
			name:    "pnpm",
			cmd:     "pnpm",
			args:    []string{"--version"},
			extract: firstLine,
		},
		{
			name: "pip",
			cmd:  "pip",
			args: []string{"--version"},
			extract: func(s string) string {
				// "pip 23.3.1 from /usr/... (python 3.11)"
				parts := strings.Fields(s)
				if len(parts) >= 2 {
					return parts[1]
				}
				return firstLine(s)
			},
		},
		{
			name: "pip3",
			cmd:  "pip3",
			args: []string{"--version"},
			extract: func(s string) string {
				parts := strings.Fields(s)
				if len(parts) >= 2 {
					return parts[1]
				}
				return firstLine(s)
			},
		},
		{
			name:    "pipx",
			cmd:     "pipx",
			args:    []string{"--version"},
			extract: firstLine,
		},
		{
			name: "poetry",
			cmd:  "poetry",
			args: []string{"--version"},
			extract: func(s string) string {
				// "Poetry (version 1.7.1)"
				s = strings.TrimPrefix(firstLine(s), "Poetry (version ")
				return strings.TrimSuffix(s, ")")
			},
		},
		{
			name: "uv",
			cmd:  "uv",
			args: []string{"--version"},
			extract: func(s string) string {
				// "uv 0.1.24 ..."
				parts := strings.Fields(s)
				if len(parts) >= 2 {
					return parts[1]
				}
				return firstLine(s)
			},
		},
		{
			name: "cargo",
			cmd:  "cargo",
			args: []string{"--version"},
			extract: func(s string) string {
				// "cargo 1.74.0 ..."
				parts := strings.Fields(s)
				if len(parts) >= 2 {
					return parts[1]
				}
				return firstLine(s)
			},
		},
		{
			name:    "gem",
			cmd:     "gem",
			args:    []string{"--version"},
			extract: firstLine,
		},
		{
			name: "bundler",
			cmd:  "bundle",
			args: []string{"--version"},
			extract: func(s string) string {
				// "Bundler version 2.4.22"
				parts := strings.Fields(s)
				if len(parts) >= 3 {
					return parts[2]
				}
				return firstLine(s)
			},
		},
		{
			name: "maven",
			cmd:  "mvn",
			args: []string{"--version"},
			extract: func(s string) string {
				// "Apache Maven 3.9.5 ..."
				for _, line := range strings.Split(s, "\n") {
					if strings.HasPrefix(line, "Apache Maven ") {
						parts := strings.Fields(line)
						if len(parts) >= 3 {
							return parts[2]
						}
					}
				}
				return firstLine(s)
			},
		},
		{
			name: "gradle",
			cmd:  "gradle",
			args: []string{"--version"},
			extract: func(s string) string {
				for _, line := range strings.Split(s, "\n") {
					if strings.HasPrefix(line, "Gradle ") {
						return strings.TrimPrefix(line, "Gradle ")
					}
				}
				return firstLine(s)
			},
		},
		{
			name: "composer",
			cmd:  "composer",
			args: []string{"--version"},
			extract: func(s string) string {
				// "Composer version 2.6.5 ..."
				parts := strings.Fields(firstLine(s))
				if len(parts) >= 3 {
					return parts[2]
				}
				return firstLine(s)
			},
		},
		{
			name: "brew",
			cmd:  "brew",
			args: []string{"--version"},
			extract: func(s string) string {
				// "Homebrew 4.1.20"
				parts := strings.Fields(firstLine(s))
				if len(parts) >= 2 {
					return parts[1]
				}
				return firstLine(s)
			},
		},
		{
			name: "apt",
			cmd:  "apt",
			args: []string{"--version"},
			extract: func(s string) string {
				parts := strings.Fields(firstLine(s))
				if len(parts) >= 2 {
					return parts[1]
				}
				return firstLine(s)
			},
		},
		{
			name:    "dnf",
			cmd:     "dnf",
			args:    []string{"--version"},
			extract: firstLine,
		},
		{
			name: "pacman",
			cmd:  "pacman",
			args: []string{"--version"},
			extract: func(s string) string {
				for _, line := range strings.Split(s, "\n") {
					if strings.Contains(line, "Pacman v") {
						parts := strings.Fields(line)
						for i, p := range parts {
							if strings.HasPrefix(p, "v") && i > 0 {
								return strings.TrimPrefix(p, "v")
							}
						}
					}
				}
				return firstLine(s)
			},
		},
		{
			name: "mix",
			cmd:  "mix",
			args: []string{"--version"},
			extract: func(s string) string {
				// "Mix 1.15.7 (compiled with Erlang/OTP 26)"
				parts := strings.Fields(firstLine(s))
				if len(parts) >= 2 {
					return parts[1]
				}
				return firstLine(s)
			},
		},
		{
			name: "nuget",
			cmd:  "nuget",
			args: []string{"help"},
			extract: func(s string) string {
				for _, line := range strings.Split(s, "\n") {
					if strings.HasPrefix(line, "NuGet Version:") {
						return strings.TrimSpace(strings.TrimPrefix(line, "NuGet Version:"))
					}
				}
				return firstLine(s)
			},
		},
		{
			name: "conan",
			cmd:  "conan",
			args: []string{"--version"},
			extract: func(s string) string {
				parts := strings.Fields(firstLine(s))
				if len(parts) >= 3 {
					return parts[2]
				}
				return firstLine(s)
			},
		},
		{
			name:    "vcpkg",
			cmd:     "vcpkg",
			args:    []string{"version"},
			extract: firstLine,
		},
		{
			name:    "go-pkg",
			cmd:     "go",
			args:    []string{"env", "GOPATH"},
			extract: firstLine,
		},
	}

	var tools []models.Tool
	seen := map[string]bool{}

	for _, d := range defs {
		if seen[d.name] {
			continue
		}
		// Skip go-pkg special case — it's metadata, not a version
		if d.name == "go-pkg" {
			continue
		}
		path, raw := runCmd(d.cmd, d.args...)
		if path == "" {
			continue
		}
		tools = append(tools, models.Tool{
			Name:    d.name,
			Version: d.extract(raw),
			Path:    path,
		})
		seen[d.name] = true
	}

	return tools
}
