package scanner

import (
	"os"
	"os/user"
	"runtime"
	"time"

	"github.com/tool-chain-brain/tcb/pkg/models"
)

const TCBVersion = "1.0.0"

// ScanSystem runs all sub-scanners and returns a fully populated Toolchain.
func ScanSystem() (*models.Toolchain, error) {
	hostname, _ := os.Hostname()
	username := ""
	if u, err := user.Current(); err == nil {
		username = u.Username
	}

	tc := &models.Toolchain{
		Meta: models.Meta{
			ScannedAt:  time.Now().UTC(),
			Hostname:   hostname,
			OS:         runtime.GOOS,
			Arch:       runtime.GOARCH,
			Username:   username,
			TCBVersion: TCBVersion,
		},
	}

	tc.Languages = ScanLanguages()
	tc.Compilers = ScanCompilers()
	tc.PackageManagers = ScanPackageManagers()
	tc.VersionManagers = ScanVersionManagers()
	tc.InfraTools = ScanInfraTools()
	tc.CrossCompilers = ScanCrossCompilers()
	tc.Git = ScanGit()
	tc.Shell = ScanShell()
	tc.Editor = ScanEditor()

	return tc, nil
}
