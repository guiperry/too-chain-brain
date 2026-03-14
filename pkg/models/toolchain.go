package models

import "time"

// Toolchain is the full snapshot of a system's developer toolchain.
type Toolchain struct {
	Meta            Meta         `yaml:"meta"`
	Languages       []Tool       `yaml:"languages,omitempty"`
	Compilers       []Tool       `yaml:"compilers,omitempty"`
	PackageManagers []Tool       `yaml:"package_managers,omitempty"`
	VersionManagers []Tool       `yaml:"version_managers,omitempty"`
	InfraTools      []Tool       `yaml:"infra_tools,omitempty"`
	CrossCompilers  []Tool       `yaml:"cross_compilers,omitempty"`
	Git             GitConfig    `yaml:"git"`
	Shell           ShellConfig  `yaml:"shell"`
	Editor          EditorConfig `yaml:"editor"`
}

// Meta holds system-level metadata captured at scan time.
type Meta struct {
	ScannedAt  time.Time `yaml:"scanned_at"`
	Hostname   string    `yaml:"hostname"`
	OS         string    `yaml:"os"`
	Arch       string    `yaml:"arch"`
	Username   string    `yaml:"username"`
	TCBVersion string    `yaml:"tcb_version"`
}

// Tool represents a single detected executable tool on the system.
type Tool struct {
	Name        string            `yaml:"name"`
	Version     string            `yaml:"version"`
	Path        string            `yaml:"path"`
	ConfigFiles []string          `yaml:"config_files,omitempty"`
	Extra       map[string]string `yaml:"extra,omitempty"`
}

// GitConfig holds values extracted from the user's global git configuration.
type GitConfig struct {
	Version          string `yaml:"version,omitempty"`
	Path             string `yaml:"path,omitempty"`
	UserName         string `yaml:"user_name,omitempty"`
	UserEmail        string `yaml:"user_email,omitempty"`
	DefaultBranch    string `yaml:"default_branch,omitempty"`
	SigningKey       string `yaml:"signing_key,omitempty"`
	GPGSign          string `yaml:"gpg_sign,omitempty"`
	CoreEditor       string `yaml:"core_editor,omitempty"`
	CoreAutoCRLF     string `yaml:"core_autocrlf,omitempty"`
	PullRebase       string `yaml:"pull_rebase,omitempty"`
	PushDefault      string `yaml:"push_default,omitempty"`
	CredentialHelper string `yaml:"credential_helper,omitempty"`
}

// ShellConfig captures the user's active shell and relevant environment.
type ShellConfig struct {
	Shell       string            `yaml:"shell"`
	Version     string            `yaml:"version,omitempty"`
	ConfigFiles []string          `yaml:"config_files,omitempty"`
	EnvVars     map[string]string `yaml:"env_vars,omitempty"`
	PathEntries []string          `yaml:"path_entries,omitempty"`
}

// EditorConfig captures detected editor configurations.
type EditorConfig struct {
	VSCode           *VSCodeConfig `yaml:"vscode,omitempty"`
	EditorConfigFile string        `yaml:"editorconfig_file,omitempty"`
}

// VSCodeConfig captures VS Code version and installed extensions.
type VSCodeConfig struct {
	Version    string   `yaml:"version,omitempty"`
	Extensions []string `yaml:"extensions,omitempty"`
}

// DiffResult holds the comparison between a saved and current toolchain.
type DiffResult struct {
	Added   []DiffEntry  `yaml:"added,omitempty"`
	Removed []DiffEntry  `yaml:"removed,omitempty"`
	Changed []DiffChange `yaml:"changed,omitempty"`
	Clean   bool         `yaml:"clean"`
}

// DiffEntry is a tool that exists in one toolchain but not the other.
type DiffEntry struct {
	Category string `yaml:"category"`
	Name     string `yaml:"name"`
	Version  string `yaml:"version"`
}

// DiffChange is a tool present in both toolchains but with a version difference.
type DiffChange struct {
	Category   string `yaml:"category"`
	Name       string `yaml:"name"`
	OldVersion string `yaml:"old_version"`
	NewVersion string `yaml:"new_version"`
}
