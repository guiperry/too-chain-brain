package main

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/tool-chain-brain/tcb/internal/export"
	"github.com/tool-chain-brain/tcb/internal/scanner"
	"github.com/tool-chain-brain/tcb/pkg/models"
)

func exportCmd() *cobra.Command {
	var outputDir string
	var inputFile string
	var noYAML bool
	var noSetupSh bool
	var noDocker bool
	var noDevCont bool

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Generate reproducible setup artifacts from a toolchain snapshot",
		Long: `Exports toolchain.yaml, setup.sh, Dockerfile.devenv, and .devcontainer/devcontainer.json.

If --input is omitted, a fresh scan is performed first.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			spin := spinner.New(spinner.CharSets[14], 80*time.Millisecond)
			spin.Color("cyan", "bold")

			var tc *models.Toolchain
			var err error

			if inputFile != "" {
				spin.Suffix = "  Loading " + inputFile + "…"
				spin.Start()
				tc, err = export.LoadYAML(inputFile)
				spin.Stop()
				if err != nil {
					return fmt.Errorf("failed to load %s: %w", inputFile, err)
				}
			} else {
				spin.Suffix = "  Scanning toolchain…"
				spin.Start()
				tc, err = scanner.ScanSystem()
				spin.Stop()
				if err != nil {
					return fmt.Errorf("scan failed: %w", err)
				}
			}

			bold := color.New(color.Bold)
			green := color.New(color.FgGreen)
			hiblack := color.New(color.FgHiBlack)

			fmt.Println()
			bold.Println("  🧠 Tool-Chain-Brain Export")
			fmt.Println("  " + hiblack.Sprint("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))

			writeStep := func(label string, fn func() (string, error)) {
				spin.Suffix = "  Writing " + label + "…"
				spin.Start()
				dest, stepErr := fn()
				spin.Stop()
				if stepErr != nil {
					fmt.Printf("  %s %-30s %s\n",
						color.RedString("✗"),
						hiblack.Sprint(label),
						color.RedString(stepErr.Error()))
					return
				}
				fmt.Printf("  %s %-30s %s\n",
					green.Sprint("✓"),
					hiblack.Sprint(label),
					color.CyanString(dest))
			}

			if !noYAML {
				writeStep("toolchain.yaml", func() (string, error) {
					return export.WriteYAML(tc, outputDir)
				})
			}
			if !noSetupSh {
				writeStep("setup.sh", func() (string, error) {
					return export.WriteSetupSh(tc, outputDir)
				})
			}
			if !noDocker {
				writeStep("Dockerfile.devenv", func() (string, error) {
					return export.WriteDockerfile(tc, outputDir)
				})
			}
			if !noDevCont {
				writeStep(".devcontainer/devcontainer.json", func() (string, error) {
					return export.WriteDevContainer(tc, outputDir)
				})
			}

			fmt.Println("  " + hiblack.Sprint("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
			fmt.Printf("\n  %s All artifacts written to %s\n\n",
				green.Sprint("✅"),
				color.CyanString(outputDir))

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output", "o", "./tcb-export", "Directory to write artifacts")
	cmd.Flags().StringVarP(&inputFile, "input", "i", "", "Use existing toolchain.yaml instead of re-scanning")
	cmd.Flags().BoolVar(&noYAML, "no-yaml", false, "Skip toolchain.yaml")
	cmd.Flags().BoolVar(&noSetupSh, "no-setup-sh", false, "Skip setup.sh")
	cmd.Flags().BoolVar(&noDocker, "no-dockerfile", false, "Skip Dockerfile.devenv")
	cmd.Flags().BoolVar(&noDevCont, "no-devcontainer", false, "Skip .devcontainer/devcontainer.json")

	return cmd
}
