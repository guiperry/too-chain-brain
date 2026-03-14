package export

import "github.com/tool-chain-brain/tcb/pkg/models"

// ToolchainRef is a thin holder used by CLI commands so they only import
// the export package rather than both export and models directly.
type ToolchainRef struct {
	TC *models.Toolchain
}
