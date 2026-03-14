package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/tool-chain-brain/tcb/internal/export"
	ghclient "github.com/tool-chain-brain/tcb/internal/github"
	"github.com/tool-chain-brain/tcb/internal/scanner"
	"github.com/tool-chain-brain/tcb/pkg/models"
)

func initRepoCmd() *cobra.Command {
	var token string
	var repoName string
	var private bool
	var inputFile string
	var skipScan bool

	cmd := &cobra.Command{
		Use:   "init-repo",
		Short: "Export toolchain artifacts and push them to a new GitHub repository",
		Long: `Scans your system (or loads an existing toolchain.yaml), generates all
artifacts, creates a GitHub repo, and commits everything in one step.

Requires a GitHub personal access token with 'repo' scope.
Set it via GITHUB_TOKEN env var or the --token flag.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			bold := color.New(color.Bold)
			green := color.New(color.FgGreen)
			hiblack := color.New(color.FgHiBlack)

			spin := spinner.New(spinner.CharSets[14], 80*time.Millisecond)
			spin.Color("cyan", "bold")

			fmt.Println()
			bold.Println("  🧠 Tool-Chain-Brain → GitHub")
			fmt.Println("  " + hiblack.Sprint("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))

			// ── Step 1: Scan or load ──────────────────────────────────────
			step := func(msg string, fn func() error) error {
				spin.Suffix = "  " + msg
				spin.Start()
				err := fn()
				spin.Stop()
				if err != nil {
					fmt.Printf("  %s %s\n", color.RedString("✗"), msg)
					return err
				}
				fmt.Printf("  %s %s\n", green.Sprint("✓"), hiblack.Sprint(msg))
				return nil
			}

			var tc *models.Toolchain
			if err := step("Scanning toolchain…", func() error {
				if inputFile != "" || skipScan {
					file := inputFile
					if file == "" {
						file = "toolchain.yaml"
					}
					loaded, err := export.LoadYAML(file)
					if err != nil {
						return fmt.Errorf("load %s: %w", file, err)
					}
					tc = loaded
					return nil
				}
				scanned, err := scanner.ScanSystem()
				if err != nil {
					return err
				}
				tc = scanned
				return nil
			}); err != nil {
				return err
			}

			// ── Step 2: Export artifacts ─────────────────────────────────
			exportDir, err := os.MkdirTemp("", "tcb-export-*")
			if err != nil {
				return fmt.Errorf("failed to create temp dir: %w", err)
			}
			defer os.RemoveAll(exportDir)

			if err := step("Generating artifacts…", func() error {
				if _, err := export.WriteYAML(tc, exportDir); err != nil {
					return err
				}
				if _, err := export.WriteSetupSh(tc, exportDir); err != nil {
					return err
				}
				if _, err := export.WriteDockerfile(tc, exportDir); err != nil {
					return err
				}
				if _, err := export.WriteDevContainer(tc, exportDir); err != nil {
					return err
				}
				return nil
			}); err != nil {
				return err
			}

			// ── Step 3: Write a README.md into the export ─────────────────
			if err := step("Writing README.md…", func() error {
				return writeRepoReadme(tc, exportDir)
			}); err != nil {
				return err
			}

			// ── Step 4: Connect to GitHub ─────────────────────────────────
			var gh *ghclient.Client
			if err := step("Authenticating with GitHub…", func() error {
				var err error
				gh, err = ghclient.NewClient(token)
				return err
			}); err != nil {
				return fmt.Errorf("%w\n\n  Set GITHUB_TOKEN or use --token", err)
			}

			// ── Step 5: Create repo ───────────────────────────────────────
			if repoName == "" {
				repoName = fmt.Sprintf("toolchain-%s", tc.Meta.Hostname)
			}

			var repoURL string
			if err := step(fmt.Sprintf("Creating repo %s/%s…", gh.Owner(), repoName), func() error {
				repo, err := gh.CreateOrGetRepo(ctx, repoName,
					fmt.Sprintf("Tool-Chain-Brain snapshot from %s (%s/%s)",
						tc.Meta.Hostname, tc.Meta.OS, tc.Meta.Arch),
					private)
				if err != nil {
					return err
				}
				repoURL = repo.GetHTMLURL()
				return nil
			}); err != nil {
				return err
			}

			// ── Step 6: Upload files ──────────────────────────────────────
			if err := step("Uploading artifacts…", func() error {
				return gh.UploadDir(ctx, repoName, exportDir, "tcb: initial toolchain snapshot")
			}); err != nil {
				return err
			}

			fmt.Println("  " + hiblack.Sprint("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
			fmt.Printf("\n  %s Repository ready: %s\n\n",
				green.Sprint("🚀"),
				color.CyanString(repoURL))

			return nil
		},
	}

	cmd.Flags().StringVar(&token, "token", "", "GitHub personal access token (or set GITHUB_TOKEN)")
	cmd.Flags().StringVar(&repoName, "repo", "", "Repository name (default: toolchain-<hostname>)")
	cmd.Flags().BoolVar(&private, "private", false, "Create repository as private")
	cmd.Flags().StringVarP(&inputFile, "input", "i", "", "Use existing toolchain.yaml instead of scanning")
	cmd.Flags().BoolVar(&skipScan, "skip-scan", false, "Use toolchain.yaml in current directory")

	return cmd
}

// writeRepoReadme writes a human-readable README into the export dir.
func writeRepoReadme(tc *models.Toolchain, dir string) error {
	content := fmt.Sprintf(`# 🧠 Tool-Chain-Brain Snapshot

> Generated by [Tool-Chain-Brain](https://github.com/tool-chain-brain/tcb) on %s

## System

| | |
|---|---|
| **Hostname** | %s |
| **OS** | %s |
| **Arch** | %s |
| **Scanned** | %s |

## Contents

| File | Description |
|---|---|
| `+"`toolchain.yaml`"+` | Full toolchain manifest |
| `+"`setup.sh`"+` | Idempotent install script |
| `+"`Dockerfile.devenv`"+` | Reproducible dev container image |
| `+"`.devcontainer/devcontainer.json`"+` | VS Code Dev Container config |

## Quick Start

### Clone and run setup
`+"```bash"+`
git clone <this-repo>
cd <this-repo>
chmod +x setup.sh && ./setup.sh
`+"```"+`

### Open in a Dev Container
1. Install [VS Code](https://code.visualstudio.com) and the [Dev Containers extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers)
2. Open this folder in VS Code
3. Click **Reopen in Container** when prompted

### Build the dev image manually
`+"```bash"+`
docker build -f Dockerfile.devenv -t devenv .
docker run -it --rm -v $(pwd):/workspace devenv
`+"```"+`

## Languages (%d)

| Name | Version |
|---|---|
`, tc.Meta.ScannedAt.Format("2006-01-02 15:04 UTC"),
		tc.Meta.Hostname,
		tc.Meta.OS,
		tc.Meta.Arch,
		tc.Meta.ScannedAt.Format("2006-01-02 15:04:05 UTC"),
		len(tc.Languages),
	)

	for _, t := range tc.Languages {
		content += fmt.Sprintf("| `%s` | %s |\n", t.Name, t.Version)
	}

	content += fmt.Sprintf(`
## Compilers — C / C++ (%d)

| Name | Version | Variant |
|---|---|---|
`, len(tc.Compilers))

	for _, t := range tc.Compilers {
		variant := ""
		if t.Extra != nil {
			variant = t.Extra["variant"]
		}
		content += fmt.Sprintf("| `%s` | %s | %s |\n", t.Name, t.Version, variant)
	}

	content += fmt.Sprintf(`
## Cross-Compilers (%d)

| Name | Version | Target / Notes |
|---|---|---|
`, len(tc.CrossCompilers))

	for _, t := range tc.CrossCompilers {
		note := ""
		if t.Extra != nil {
			if v, ok := t.Extra["target"]; ok {
				note = v
			} else if v, ok := t.Extra["macos_sdk"]; ok {
				note = "macOS SDK " + v
			} else if v, ok := t.Extra["note"]; ok {
				note = v
			}
		}
		content += fmt.Sprintf("| `%s` | %s | %s |\n", t.Name, t.Version, note)
	}

	content += fmt.Sprintf(`
## Package Managers (%d)

| Name | Version |
|---|---|
`, len(tc.PackageManagers))

	for _, t := range tc.PackageManagers {
		content += fmt.Sprintf("| `%s` | %s |\n", t.Name, t.Version)
	}

	content += fmt.Sprintf(`
## Infrastructure Tools (%d)

| Name | Version |
|---|---|
`, len(tc.InfraTools))

	for _, t := range tc.InfraTools {
		content += fmt.Sprintf("| `%s` | %s |\n", t.Name, t.Version)
	}

	content += `
---
*Snapshot managed by [Tool-Chain-Brain](https://github.com/tool-chain-brain/tcb)*
`

	return os.WriteFile(filepath.Join(dir, "README.md"), []byte(content), 0644)
}
