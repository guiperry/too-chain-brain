package diff

import (
	"github.com/tool-chain-brain/tcb/pkg/models"
)

// Compare produces a DiffResult between a saved (old) toolchain and the current (new) one.
func Compare(old, current *models.Toolchain) models.DiffResult {
	result := models.DiffResult{}

	categories := map[string]struct {
		old []models.Tool
		cur []models.Tool
	}{
		"languages":        {old.Languages, current.Languages},
		"compilers":        {old.Compilers, current.Compilers},
		"package_managers": {old.PackageManagers, current.PackageManagers},
		"version_managers": {old.VersionManagers, current.VersionManagers},
		"infra_tools":      {old.InfraTools, current.InfraTools},
		"cross_compilers":  {old.CrossCompilers, current.CrossCompilers},
	}

	for category, pair := range categories {
		oldMap := toolMap(pair.old)
		curMap := toolMap(pair.cur)

		for name, curTool := range curMap {
			if oldTool, exists := oldMap[name]; !exists {
				// New tool — added
				result.Added = append(result.Added, models.DiffEntry{
					Category: category,
					Name:     name,
					Version:  curTool.Version,
				})
			} else if oldTool.Version != curTool.Version {
				// Tool exists in both but version changed
				result.Changed = append(result.Changed, models.DiffChange{
					Category:   category,
					Name:       name,
					OldVersion: oldTool.Version,
					NewVersion: curTool.Version,
				})
			}
		}

		for name, oldTool := range oldMap {
			if _, exists := curMap[name]; !exists {
				// Tool removed
				result.Removed = append(result.Removed, models.DiffEntry{
					Category: category,
					Name:     name,
					Version:  oldTool.Version,
				})
			}
		}
	}

	result.Clean = len(result.Added) == 0 && len(result.Removed) == 0 && len(result.Changed) == 0
	return result
}

func toolMap(tools []models.Tool) map[string]models.Tool {
	m := make(map[string]models.Tool, len(tools))
	for _, t := range tools {
		m[t.Name] = t
	}
	return m
}
