package scanner

import (
	"strings"

	"github.com/tool-chain-brain/tcb/pkg/models"
)

type langDef struct {
	name    string
	cmd     string
	args    []string
	extract func(string) string
}

// ScanLanguages probes all known language runtimes and returns those found.
func ScanLanguages() []models.Tool {
	defs := []langDef{
		{
			name: "go",
			cmd:  "go",
			args: []string{"version"},
			extract: func(s string) string {
				// "go version go1.21.0 linux/amd64"
				parts := strings.Fields(s)
				if len(parts) >= 3 {
					return strings.TrimPrefix(parts[2], "go")
				}
				return s
			},
		},
		{
			name: "node",
			cmd:  "node",
			args: []string{"--version"},
			extract: func(s string) string {
				return strings.TrimPrefix(firstLine(s), "v")
			},
		},
		{
			name: "python3",
			cmd:  "python3",
			args: []string{"--version"},
			extract: func(s string) string {
				return strings.TrimPrefix(firstLine(s), "Python ")
			},
		},
		{
			name: "python",
			cmd:  "python",
			args: []string{"--version"},
			extract: func(s string) string {
				return strings.TrimPrefix(firstLine(s), "Python ")
			},
		},
		{
			name: "ruby",
			cmd:  "ruby",
			args: []string{"--version"},
			extract: func(s string) string {
				// "ruby 3.2.0 (2022-12-25 revision ...)"
				parts := strings.Fields(s)
				if len(parts) >= 2 {
					return parts[1]
				}
				return firstLine(s)
			},
		},
		{
			name: "rustc",
			cmd:  "rustc",
			args: []string{"--version"},
			extract: func(s string) string {
				// "rustc 1.74.0 (79e9716c9 2023-11-13)"
				parts := strings.Fields(s)
				if len(parts) >= 2 {
					return parts[1]
				}
				return firstLine(s)
			},
		},
		{
			name: "java",
			cmd:  "java",
			args: []string{"-version"},
			extract: func(s string) string {
				// 'openjdk version "21.0.1" 2023-10-17'
				if idx := strings.Index(s, "\""); idx != -1 {
					rest := s[idx+1:]
					if end := strings.Index(rest, "\""); end != -1 {
						return rest[:end]
					}
				}
				return firstLine(s)
			},
		},
		{
			name: "php",
			cmd:  "php",
			args: []string{"--version"},
			extract: func(s string) string {
				// "PHP 8.2.0 (cli) ..."
				parts := strings.Fields(firstLine(s))
				if len(parts) >= 2 {
					return parts[1]
				}
				return firstLine(s)
			},
		},
		{
			name:    "dotnet",
			cmd:     "dotnet",
			args:    []string{"--version"},
			extract: firstLine,
		},
		{
			name: "swift",
			cmd:  "swift",
			args: []string{"--version"},
			extract: func(s string) string {
				// "Apple Swift version 5.9.2 ..."
				if idx := strings.Index(s, "Swift version "); idx != -1 {
					rest := s[idx+14:]
					parts := strings.Fields(rest)
					if len(parts) > 0 {
						return parts[0]
					}
				}
				return firstLine(s)
			},
		},
		{
			name: "kotlinc",
			cmd:  "kotlinc",
			args: []string{"-version"},
			extract: func(s string) string {
				// "kotlinc-jvm 1.9.21 ..."
				parts := strings.Fields(s)
				if len(parts) >= 2 {
					return parts[1]
				}
				return firstLine(s)
			},
		},
		{
			name: "scala",
			cmd:  "scala",
			args: []string{"-version"},
			extract: func(s string) string {
				parts := strings.Fields(s)
				for i, p := range parts {
					if strings.EqualFold(p, "version") && i+1 < len(parts) {
						return parts[i+1]
					}
				}
				return firstLine(s)
			},
		},
		{
			name: "elixir",
			cmd:  "elixir",
			args: []string{"--version"},
			extract: func(s string) string {
				for _, line := range strings.Split(s, "\n") {
					if strings.HasPrefix(line, "Elixir ") {
						parts := strings.Fields(line)
						if len(parts) >= 2 {
							return parts[1]
						}
					}
				}
				return firstLine(s)
			},
		},
		{
			name: "deno",
			cmd:  "deno",
			args: []string{"--version"},
			extract: func(s string) string {
				for _, line := range strings.Split(s, "\n") {
					if strings.HasPrefix(line, "deno ") {
						return strings.TrimPrefix(line, "deno ")
					}
				}
				return firstLine(s)
			},
		},
		{
			name:    "bun",
			cmd:     "bun",
			args:    []string{"--version"},
			extract: firstLine,
		},
		{
			name:    "zig",
			cmd:     "zig",
			args:    []string{"version"},
			extract: firstLine,
		},
		{
			name: "lua",
			cmd:  "lua",
			args: []string{"-v"},
			extract: func(s string) string {
				// "Lua 5.4.6 ..."
				parts := strings.Fields(s)
				if len(parts) >= 2 {
					return parts[1]
				}
				return firstLine(s)
			},
		},
		{
			name: "perl",
			cmd:  "perl",
			args: []string{"--version"},
			extract: func(s string) string {
				// "This is perl 5, version 36, ..."
				if idx := strings.Index(s, "version "); idx != -1 {
					rest := s[idx+8:]
					parts := strings.Fields(rest)
					if len(parts) > 0 {
						return "5." + strings.TrimSuffix(parts[0], ",")
					}
				}
				return firstLine(s)
			},
		},
		{
			name: "dart",
			cmd:  "dart",
			args: []string{"--version"},
			extract: func(s string) string {
				// "Dart SDK version: 3.2.0 ..."
				if idx := strings.Index(s, "version: "); idx != -1 {
					rest := s[idx+9:]
					parts := strings.Fields(rest)
					if len(parts) > 0 {
						return parts[0]
					}
				}
				return firstLine(s)
			},
		},
	}

	var tools []models.Tool
	seen := map[string]bool{}

	for _, d := range defs {
		if seen[d.name] {
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
