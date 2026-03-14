package main

import (
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/tool-chain-brain/tcb/internal/diff"
	"github.com/tool-chain-brain/tcb/internal/export"
	"github.com/tool-chain-brain/tcb/internal/scanner"
)

func diffCmd() *cobra.Command {
	var currentFile string

	cmd := &cobra.Command{
		Use:   "diff <toolchain.yaml>",
		Short: "Compare a saved toolchain snapshot against the current system",
		Long: `Loads a previously saved toolchain.yaml and compares it against
a fresh scan of the current system (or another yaml via --current).

Exit code 0 = clean (no differences)
Exit code 1 = differences found
Exit code 2 = error`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			savedPath := args[0]

			spin := spinner.New(spinner.CharSets[14], 80*time.Millisecond)
			spin.Color("cyan", "bold")

			// Load saved snapshot
			spin.Suffix = "  Loading " + savedPath + "…"
			spin.Start()
			saved, err := export.LoadYAML(savedPath)
			spin.Stop()
			if err != nil {
				return fmt.Errorf("failed to load %s: %w", savedPath, err)
			}

			// Load or scan current state
			var currentTC = saved // default to comparing with itself (will be replaced)
			if currentFile != "" {
				spin.Suffix = "  Loading " + currentFile + "…"
				spin.Start()
				currentTC, err = export.LoadYAML(currentFile)
				spin.Stop()
				if err != nil {
					return fmt.Errorf("failed to load %s: %w", currentFile, err)
				}
			} else {
				spin.Suffix = "  Scanning current system…"
				spin.Start()
				currentTC, err = scanner.ScanSystem()
				spin.Stop()
				if err != nil {
					return fmt.Errorf("scan failed: %w", err)
				}
			}

			result := diff.Compare(saved, currentTC)

			bold := color.New(color.Bold)
			hiblack := color.New(color.FgHiBlack)

			fmt.Println()
			bold.Println("  🧠 Tool-Chain-Brain Diff")
			fmt.Println("  " + hiblack.Sprint("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
			fmt.Printf("  %s  %s\n", hiblack.Sprint("Saved:  "), savedPath)
			if currentFile != "" {
				fmt.Printf("  %s  %s\n", hiblack.Sprint("Current:"), currentFile)
			} else {
				fmt.Printf("  %s  %s\n", hiblack.Sprint("Current:"), "live system scan")
			}
			fmt.Println("  " + hiblack.Sprint("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
			fmt.Println()

			if result.Clean {
				color.Green("  ✅  No differences — system matches saved toolchain\n")
				return nil
			}

			if len(result.Added) > 0 {
				color.Green("  ➕  Added (%d)\n", len(result.Added))
				for _, e := range result.Added {
					fmt.Printf("     %s  %s  %s  %s\n",
						color.GreenString("+"),
						color.CyanString("%-20s", e.Name),
						hiblack.Sprintf("%-14s", e.Version),
						hiblack.Sprintf("[%s]", e.Category))
				}
				fmt.Println()
			}

			if len(result.Removed) > 0 {
				color.Red("  ➖  Removed (%d)\n", len(result.Removed))
				for _, e := range result.Removed {
					fmt.Printf("     %s  %s  %s  %s\n",
						color.RedString("-"),
						color.CyanString("%-20s", e.Name),
						hiblack.Sprintf("%-14s", e.Version),
						hiblack.Sprintf("[%s]", e.Category))
				}
				fmt.Println()
			}

			if len(result.Changed) > 0 {
				color.Yellow("  🔄  Changed (%d)\n", len(result.Changed))
				for _, c := range result.Changed {
					fmt.Printf("     %s  %s  %s → %s  %s\n",
						color.YellowString("~"),
						color.CyanString("%-20s", c.Name),
						color.RedString("%-14s", c.OldVersion),
						color.GreenString("%-14s", c.NewVersion),
						hiblack.Sprintf("[%s]", c.Category))
				}
				fmt.Println()
			}

			// Non-zero exit when differences exist (useful in CI)
			os.Exit(1)
			return nil
		},
	}

	cmd.Flags().StringVar(&currentFile, "current", "", "Compare against another toolchain.yaml instead of live scan")

	return cmd
}
