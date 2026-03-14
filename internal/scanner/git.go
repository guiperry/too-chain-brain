package scanner

import (
	"strings"

	"github.com/tool-chain-brain/tcb/pkg/models"
)

// ScanGit reads global git configuration and git version.
func ScanGit() models.GitConfig {
	cfg := models.GitConfig{}

	path, raw := runCmd("git", "--version")
	if path == "" {
		return cfg
	}

	// "git version 2.43.0"
	parts := strings.Fields(raw)
	if len(parts) >= 3 {
		cfg.Version = parts[2]
	}
	cfg.Path = path

	cfg.UserName = gitConfigValue("user.name")
	cfg.UserEmail = gitConfigValue("user.email")
	cfg.DefaultBranch = gitConfigValue("init.defaultBranch")
	cfg.SigningKey = gitConfigValue("user.signingkey")
	cfg.GPGSign = gitConfigValue("commit.gpgsign")
	cfg.CoreEditor = gitConfigValue("core.editor")
	cfg.CoreAutoCRLF = gitConfigValue("core.autocrlf")
	cfg.PullRebase = gitConfigValue("pull.rebase")
	cfg.PushDefault = gitConfigValue("push.default")
	cfg.CredentialHelper = gitConfigValue("credential.helper")

	return cfg
}
