package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tcb",
	Short: "Tool-Chain-Brain — snapshot, replicate, and diff developer toolchains",
	Long: color.CyanString(`
 ████████╗ ██████╗██████╗
    ██╔══╝██╔════╝██╔══██╗
    ██║   ██║     ██████╔╝
    ██║   ██║     ██╔══██╗
    ██║   ╚██████╗██████╔╝
    ╚═╝    ╚═════╝╚═════╝
 Tool-Chain-Brain`) + `

Scan your developer toolchain, export reproducible setup artifacts,
diff environments, and push everything to a GitHub repo — all in one command.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func main() {
	rootCmd.AddCommand(scanCmd())
	rootCmd.AddCommand(exportCmd())
	rootCmd.AddCommand(diffCmd())
	rootCmd.AddCommand(initRepoCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, color.RedString("error: %s", err))
		os.Exit(1)
	}
}
