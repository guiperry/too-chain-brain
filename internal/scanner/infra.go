package scanner

import (
	"strings"

	"github.com/tool-chain-brain/tcb/pkg/models"
)

// ScanInfraTools probes container, cloud, and infrastructure tooling.
func ScanInfraTools() []models.Tool {
	type infraDef struct {
		name    string
		cmd     string
		args    []string
		extract func(string) string
	}

	defs := []infraDef{
		// Container tooling
		{
			name: "docker",
			cmd:  "docker",
			args: []string{"--version"},
			extract: func(s string) string {
				// "Docker version 24.0.7, build afdd53b4e3"
				parts := strings.Fields(s)
				if len(parts) >= 3 {
					return strings.TrimSuffix(parts[2], ",")
				}
				return firstLine(s)
			},
		},
		{
			name: "docker-compose",
			cmd:  "docker-compose",
			args: []string{"--version"},
			extract: func(s string) string {
				// "Docker Compose version v2.23.0" or older "docker-compose version 1.29.2"
				parts := strings.Fields(s)
				for _, p := range parts {
					if strings.HasPrefix(p, "v") || (len(p) > 0 && p[0] >= '0' && p[0] <= '9') {
						return strings.TrimPrefix(p, "v")
					}
				}
				return firstLine(s)
			},
		},
		{
			name: "podman",
			cmd:  "podman",
			args: []string{"--version"},
			extract: func(s string) string {
				// "podman version 4.7.2"
				parts := strings.Fields(s)
				if len(parts) >= 3 {
					return parts[2]
				}
				return firstLine(s)
			},
		},
		{
			name: "nerdctl",
			cmd:  "nerdctl",
			args: []string{"--version"},
			extract: func(s string) string {
				parts := strings.Fields(s)
				if len(parts) >= 3 {
					return parts[2]
				}
				return firstLine(s)
			},
		},
		// Kubernetes
		{
			name: "kubectl",
			cmd:  "kubectl",
			args: []string{"version", "--client", "--short"},
			extract: func(s string) string {
				// "Client Version: v1.28.3"
				for _, line := range strings.Split(s, "\n") {
					if strings.Contains(line, "Client Version:") {
						parts := strings.Fields(line)
						if len(parts) >= 3 {
							return strings.TrimPrefix(parts[2], "v")
						}
					}
				}
				// Newer kubectl: just version string
				return strings.TrimPrefix(firstLine(s), "v")
			},
		},
		{
			name: "helm",
			cmd:  "helm",
			args: []string{"version", "--short"},
			extract: func(s string) string {
				// "v3.13.2+g979d061"
				parts := strings.SplitN(firstLine(s), "+", 2)
				return strings.TrimPrefix(parts[0], "v")
			},
		},
		{
			name: "minikube",
			cmd:  "minikube",
			args: []string{"version"},
			extract: func(s string) string {
				for _, line := range strings.Split(s, "\n") {
					if strings.HasPrefix(line, "minikube version:") {
						return strings.TrimPrefix(strings.TrimSpace(strings.TrimPrefix(line, "minikube version:")), "v")
					}
				}
				return firstLine(s)
			},
		},
		{
			name: "k3s",
			cmd:  "k3s",
			args: []string{"--version"},
			extract: func(s string) string {
				// "k3s version v1.28.3+k3s2 ..."
				parts := strings.Fields(firstLine(s))
				if len(parts) >= 3 {
					return strings.TrimPrefix(strings.SplitN(parts[2], "+", 2)[0], "v")
				}
				return firstLine(s)
			},
		},
		{
			name: "kind",
			cmd:  "kind",
			args: []string{"--version"},
			extract: func(s string) string {
				// "kind v0.20.0 go1.21.0 linux/amd64"
				parts := strings.Fields(s)
				if len(parts) >= 2 {
					return strings.TrimPrefix(parts[1], "v")
				}
				return firstLine(s)
			},
		},
		// Infrastructure as Code
		{
			name: "terraform",
			cmd:  "terraform",
			args: []string{"version"},
			extract: func(s string) string {
				// "Terraform v1.6.5\n..."
				line := firstLine(s)
				return strings.TrimPrefix(line, "Terraform v")
			},
		},
		{
			name: "tofu",
			cmd:  "tofu",
			args: []string{"version"},
			extract: func(s string) string {
				line := firstLine(s)
				return strings.TrimPrefix(line, "OpenTofu v")
			},
		},
		{
			name: "pulumi",
			cmd:  "pulumi",
			args: []string{"version"},
			extract: func(s string) string {
				return strings.TrimPrefix(firstLine(s), "v")
			},
		},
		{
			name: "ansible",
			cmd:  "ansible",
			args: []string{"--version"},
			extract: func(s string) string {
				// "ansible [core 2.15.6]\n..."
				line := firstLine(s)
				if idx := strings.Index(line, "[core "); idx != -1 {
					rest := line[idx+6:]
					return strings.TrimSuffix(rest, "]")
				}
				// Older: "ansible 2.9.x"
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					return parts[1]
				}
				return line
			},
		},
		{
			name:    "packer",
			cmd:     "packer",
			args:    []string{"--version"},
			extract: firstLine,
		},
		{
			name: "vault",
			cmd:  "vault",
			args: []string{"--version"},
			extract: func(s string) string {
				// "Vault v1.15.2 ..."
				parts := strings.Fields(firstLine(s))
				if len(parts) >= 2 {
					return strings.TrimPrefix(parts[1], "v")
				}
				return firstLine(s)
			},
		},
		// Cloud CLIs
		{
			name: "aws",
			cmd:  "aws",
			args: []string{"--version"},
			extract: func(s string) string {
				// "aws-cli/2.15.0 ..." or "aws-cli/1.x..."
				for _, field := range strings.Fields(s) {
					if strings.HasPrefix(field, "aws-cli/") {
						return strings.TrimPrefix(field, "aws-cli/")
					}
				}
				return firstLine(s)
			},
		},
		{
			name: "gcloud",
			cmd:  "gcloud",
			args: []string{"--version"},
			extract: func(s string) string {
				// "Google Cloud SDK 456.0.0\n..."
				for _, line := range strings.Split(s, "\n") {
					if strings.HasPrefix(line, "Google Cloud SDK ") {
						return strings.TrimPrefix(line, "Google Cloud SDK ")
					}
				}
				return firstLine(s)
			},
		},
		{
			name: "az",
			cmd:  "az",
			args: []string{"--version"},
			extract: func(s string) string {
				// "azure-cli 2.55.0\n..."
				for _, line := range strings.Split(s, "\n") {
					if strings.HasPrefix(line, "azure-cli") {
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
			name: "doctl",
			cmd:  "doctl",
			args: []string{"version"},
			extract: func(s string) string {
				// "doctl version 1.101.0-release ..."
				parts := strings.Fields(firstLine(s))
				if len(parts) >= 3 {
					return strings.SplitN(parts[2], "-", 2)[0]
				}
				return firstLine(s)
			},
		},
		// VCS & Collaboration
		{
			name: "gh",
			cmd:  "gh",
			args: []string{"--version"},
			extract: func(s string) string {
				// "gh version 2.39.2 ..."
				parts := strings.Fields(firstLine(s))
				if len(parts) >= 3 {
					return parts[2]
				}
				return firstLine(s)
			},
		},
		// Build & task tools
		{
			name: "make",
			cmd:  "make",
			args: []string{"--version"},
			extract: func(s string) string {
				// "GNU Make 4.3\n..."
				line := firstLine(s)
				if strings.HasPrefix(line, "GNU Make ") {
					return strings.TrimPrefix(line, "GNU Make ")
				}
				return line
			},
		},
		{
			name:    "just",
			cmd:     "just",
			args:    []string{"--version"},
			extract: firstLine,
		},
		{
			name: "task",
			cmd:  "task",
			args: []string{"--version"},
			extract: func(s string) string {
				// "Task version: v3.31.0"
				if idx := strings.Index(s, "v3."); idx != -1 {
					return s[idx+1:]
				}
				return firstLine(s)
			},
		},
		{
			name: "bazel",
			cmd:  "bazel",
			args: []string{"--version"},
			extract: func(s string) string {
				// "bazel 6.4.0"
				parts := strings.Fields(firstLine(s))
				if len(parts) >= 2 {
					return parts[1]
				}
				return firstLine(s)
			},
		},
		{
			name: "cmake",
			cmd:  "cmake",
			args: []string{"--version"},
			extract: func(s string) string {
				// "cmake version 3.27.7"
				parts := strings.Fields(firstLine(s))
				if len(parts) >= 3 {
					return parts[2]
				}
				return firstLine(s)
			},
		},
		// CI/CD
		{
			name: "act",
			cmd:  "act",
			args: []string{"--version"},
			extract: func(s string) string {
				// "act version 0.2.60"
				parts := strings.Fields(firstLine(s))
				if len(parts) >= 3 {
					return parts[2]
				}
				return firstLine(s)
			},
		},
	}

	var tools []models.Tool
	for _, d := range defs {
		path, raw := runCmd(d.cmd, d.args...)
		if path == "" {
			continue
		}
		tools = append(tools, models.Tool{
			Name:    d.name,
			Version: d.extract(raw),
			Path:    path,
		})
	}

	return tools
}
