package main

import (
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/tool-chain-brain/tcb/internal/export"
	"github.com/tool-chain-brain/tcb/internal/scanner"
)

func scanCmd() *cobra.Command {
	var outputDir string
	var noSave bool
	var quiet bool

	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan the system's developer toolchain",
		Long:  "Probes all installed language runtimes, package managers, version managers, infrastructure tools, git config, shell, and editor settings.",
		RunE: func(cmd *cobra.Command, args []string) error {
			spin := spinner.New(spinner.CharSets[14], 80*time.Millisecond)
			if !quiet {
				spin.Suffix = "  Scanning toolchain…"
				spin.Color("cyan", "bold")
				spin.Start()
			}

			tc, err := scanner.ScanSystem()
			if !quiet {
				spin.Stop()
			}
			if err != nil {
				return fmt.Errorf("scan failed: %w", err)
			}

			if !quiet {
				printScanSummary(tc.Meta.Hostname, tc.Meta.OS, tc.Meta.Arch,
					len(tc.Languages), len(tc.PackageManagers),
					len(tc.VersionManagers), len(tc.InfraTools))
			}

			if !noSave {
				dest, err := export.WriteYAML(tc, outputDir)
				if err != nil {
					return fmt.Errorf("failed to write toolchain.yaml: %w", err)
				}
				if !quiet {
					color.Green("  ✓ Saved → %s\n", dest)
				} else {
					fmt.Println(dest)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output", "o", ".", "Directory to write toolchain.yaml")
	cmd.Flags().BoolVar(&noSave, "dry-run", false, "Print summary without saving toolchain.yaml")
	cmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Only output the saved file path")

	return cmd
}

func printScanSummary(hostname, goos, arch string, langs, pkgMgrs, verMgrs, infra int) {
	bold := color.New(color.Bold)
	cyan := color.New(color.FgCyan)

	fmt.Println()
	bold.Println("  🧠 Tool-Chain-Brain Scan Results")
	fmt.Println("  " + color.HiBlackString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Printf("  %-22s %s\n", color.HiBlackString("Host:"), hostname)
	fmt.Printf("  %-22s %s/%s\n", color.HiBlackString("Platform:"), goos, arch)
	fmt.Println("  " + color.HiBlackString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))

	printStat := func(label string, count int) {
		icon := color.GreenString("✓")
		if count == 0 {
			icon = color.HiBlackString("·")
		}
		fmt.Printf("  %s %-22s %s\n", icon, color.HiBlackString(label), cyan.Sprintf("%d found", count))
	}

	printStat("Language runtimes", langs)
	printStat("Package managers", pkgMgrs)
	printStat("Version managers", verMgrs)
	printStat("Infra / DevOps tools", infra)

	fmt.Println()

	// Warn if nothing was found at all
	total := langs + pkgMgrs + verMgrs + infra
	if total == 0 {
		_, _ = fmt.Fprintln(os.Stderr, color.YellowString("  ⚠  No tools detected. Are they on your PATH?"))
	}
}
