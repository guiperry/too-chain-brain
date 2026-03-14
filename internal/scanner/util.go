package scanner

import (
	"os/exec"
	"strings"
)

// runCmd executes a command and returns its combined stdout+stderr output.
// Returns empty strings if the binary is not found.
func runCmd(cmd string, args ...string) (path string, output string) {
	p, err := exec.LookPath(cmd)
	if err != nil {
		return "", ""
	}
	out, _ := exec.Command(cmd, args...).CombinedOutput()
	return p, strings.TrimSpace(string(out))
}

// firstLine returns the first non-empty line of s.
func firstLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return strings.TrimSpace(s)
}

// gitConfigValue reads a single git config value globally.
func gitConfigValue(key string) string {
	_, err := exec.LookPath("git")
	if err != nil {
		return ""
	}
	out, err := exec.Command("git", "config", "--global", key).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
