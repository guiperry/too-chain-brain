package scanner

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/tool-chain-brain/tcb/pkg/models"
)

// ScanCompilers detects native C/C++ compilers on the system.
func ScanCompilers() []models.Tool {
	type compilerDef struct {
		name    string
		cmd     string
		args    []string
		extract func(string) string
	}

	defs := []compilerDef{
		// GCC family
		{
			name: "gcc",
			cmd:  "gcc",
			args: []string{"--version"},
			extract: func(s string) string {
				// "gcc (Ubuntu 13.2.0-4ubuntu3) 13.2.0"
				// "Apple clang version 15.0.0 ..." (macOS alias)
				// We want the version number on the first line.
				line := firstLine(s)
				parts := strings.Fields(line)
				// Version is typically the last field on the first line
				if len(parts) > 0 {
					return parts[len(parts)-1]
				}
				return line
			},
		},
		{
			name: "gcc-13",
			cmd:  "gcc-13",
			args: []string{"--version"},
			extract: func(s string) string {
				parts := strings.Fields(firstLine(s))
				if len(parts) > 0 {
					return parts[len(parts)-1]
				}
				return firstLine(s)
			},
		},
		{
			name: "gcc-12",
			cmd:  "gcc-12",
			args: []string{"--version"},
			extract: func(s string) string {
				parts := strings.Fields(firstLine(s))
				if len(parts) > 0 {
					return parts[len(parts)-1]
				}
				return firstLine(s)
			},
		},
		{
			name: "g++",
			cmd:  "g++",
			args: []string{"--version"},
			extract: func(s string) string {
				parts := strings.Fields(firstLine(s))
				if len(parts) > 0 {
					return parts[len(parts)-1]
				}
				return firstLine(s)
			},
		},
		// Clang family
		{
			name: "clang",
			cmd:  "clang",
			args: []string{"--version"},
			extract: func(s string) string {
				// "Apple clang version 15.0.0 (clang-1500.3.9.4)"
				// "clang version 17.0.6"
				line := firstLine(s)
				if idx := strings.Index(line, "version "); idx != -1 {
					rest := strings.Fields(line[idx+8:])
					if len(rest) > 0 {
						return rest[0]
					}
				}
				return line
			},
		},
		{
			name: "clang++",
			cmd:  "clang++",
			args: []string{"--version"},
			extract: func(s string) string {
				line := firstLine(s)
				if idx := strings.Index(line, "version "); idx != -1 {
					rest := strings.Fields(line[idx+8:])
					if len(rest) > 0 {
						return rest[0]
					}
				}
				return line
			},
		},
		// LLVM tools
		{
			name: "llvm-gcc",
			cmd:  "llvm-gcc",
			args: []string{"--version"},
			extract: func(s string) string {
				parts := strings.Fields(firstLine(s))
				if len(parts) > 0 {
					return parts[len(parts)-1]
				}
				return firstLine(s)
			},
		},
		// Generic cc / c++ aliases — capture what they resolve to
		{
			name: "cc",
			cmd:  "cc",
			args: []string{"--version"},
			extract: func(s string) string {
				line := firstLine(s)
				if idx := strings.Index(line, "version "); idx != -1 {
					rest := strings.Fields(line[idx+8:])
					if len(rest) > 0 {
						return rest[0]
					}
				}
				parts := strings.Fields(line)
				if len(parts) > 0 {
					return parts[len(parts)-1]
				}
				return line
			},
		},
		{
			name: "c++",
			cmd:  "c++",
			args: []string{"--version"},
			extract: func(s string) string {
				line := firstLine(s)
				if idx := strings.Index(line, "version "); idx != -1 {
					rest := strings.Fields(line[idx+8:])
					if len(rest) > 0 {
						return rest[0]
					}
				}
				return line
			},
		},
		// GNU assembler / linker (often ship alongside gcc)
		{
			name: "as",
			cmd:  "as",
			args: []string{"--version"},
			extract: func(s string) string {
				// "GNU assembler version 2.41"
				line := firstLine(s)
				if idx := strings.Index(line, "version "); idx != -1 {
					rest := strings.Fields(line[idx+8:])
					if len(rest) > 0 {
						return rest[0]
					}
				}
				return line
			},
		},
		{
			name: "ld",
			cmd:  "ld",
			args: []string{"--version"},
			extract: func(s string) string {
				// "GNU ld (GNU Binutils) 2.41"
				line := firstLine(s)
				parts := strings.Fields(line)
				if len(parts) > 0 {
					return parts[len(parts)-1]
				}
				return line
			},
		},
		// NASM / YASM assemblers
		{
			name: "nasm",
			cmd:  "nasm",
			args: []string{"--version"},
			extract: func(s string) string {
				// "NASM version 2.16.01 compiled on ..."
				parts := strings.Fields(firstLine(s))
				if len(parts) >= 3 {
					return parts[2]
				}
				return firstLine(s)
			},
		},
		{
			name: "yasm",
			cmd:  "yasm",
			args: []string{"--version"},
			extract: func(s string) string {
				// "yasm 1.3.0"
				parts := strings.Fields(firstLine(s))
				if len(parts) >= 2 {
					return parts[1]
				}
				return firstLine(s)
			},
		},
		// musl libc compiler wrapper
		{
			name: "musl-gcc",
			cmd:  "musl-gcc",
			args: []string{"--version"},
			extract: func(s string) string {
				parts := strings.Fields(firstLine(s))
				if len(parts) > 0 {
					return parts[len(parts)-1]
				}
				return firstLine(s)
			},
		},
	}

	var tools []models.Tool
	seen := map[string]bool{}

	for _, d := range defs {
		if seen[d.name] {
			continue
		}
		path, raw := runCmd(d.cmd, d.args...)
		if path == "" {
			continue
		}
		extra := map[string]string{}

		// On macOS, tag whether gcc/clang is the Apple wrapper or real GCC
		if strings.Contains(raw, "Apple clang") {
			extra["variant"] = "Apple clang (Xcode)"
		} else if strings.Contains(raw, "Apple LLVM") {
			extra["variant"] = "Apple LLVM"
		} else if strings.Contains(strings.ToLower(raw), "free software foundation") ||
			strings.Contains(raw, "GCC") {
			extra["variant"] = "GNU GCC"
		}

		tool := models.Tool{
			Name:    d.name,
			Version: d.extract(raw),
			Path:    path,
		}
		if len(extra) > 0 {
			tool.Extra = extra
		}

		tools = append(tools, tool)
		seen[d.name] = true
	}

	return tools
}

// ScanCrossCompilers detects cross-compilation toolchains:
// osxcross (macOS target), mingw-w64 (Windows target), and Wine.
func ScanCrossCompilers() []models.Tool {
	var tools []models.Tool

	// ── osxcross ─────────────────────────────────────────────────────────────
	// osxcross installs a set of clang/gcc wrappers named like:
	//   o64-clang, o64-clang++, o32-clang
	//   x86_64-apple-darwin*-clang, arm64-apple-darwin*-clang
	// and sets OSXCROSS_ROOT to the install directory.

	osxcrossRoot := os.Getenv("OSXCROSS_ROOT")
	if osxcrossRoot == "" {
		// Common installation directories for osxcross-master
		home, _ := os.UserHomeDir()
		candidates := []string{
			filepath.Join(home, "osxcross"),
			filepath.Join(home, "osxcross-master"),
			"/usr/local/osxcross",
			"/opt/osxcross",
			"/usr/osxcross",
			"/usr/local/osxcross-master",
		}
		for _, c := range candidates {
			if stat, err := os.Stat(c); err == nil && stat.IsDir() {
				osxcrossRoot = c
				break
			}
		}
	}

	// Also probe for the wrappers themselves on PATH even if root not found
	osxcrossWrappers := []string{
		"o64-clang", "o64-clang++",
		"o32-clang", "o32-clang++",
		"oa64-clang", "oa64-clang++",
	}

	// Search for versioned Darwin wrappers on PATH (e.g. x86_64-apple-darwin23-clang)
	darwinWrappers := findDarwinWrappersOnPath()
	osxcrossWrappers = append(osxcrossWrappers, darwinWrappers...)

	foundOsxcross := false
	for _, wrapper := range osxcrossWrappers {
		path, raw := runCmd(wrapper, "--version")
		if path == "" {
			continue
		}
		version := ""
		if idx := strings.Index(raw, "version "); idx != -1 {
			rest := strings.Fields(raw[idx+8:])
			if len(rest) > 0 {
				version = rest[0]
			}
		}

		extra := map[string]string{"wrapper": wrapper}
		if osxcrossRoot != "" {
			extra["osxcross_root"] = osxcrossRoot
		}

		// Detect target SDK from wrapper name or osxcross directory
		if sdk := detectOsxcrossSDK(osxcrossRoot); sdk != "" {
			extra["macos_sdk"] = sdk
		}

		tools = append(tools, models.Tool{
			Name:    "osxcross",
			Version: version,
			Path:    path,
			Extra:   extra,
		})
		foundOsxcross = true
		break // report once; additional wrappers are all from same install
	}

	// If we found the root dir but no wrappers on PATH, still report it
	if !foundOsxcross && osxcrossRoot != "" {
		extra := map[string]string{"osxcross_root": osxcrossRoot}
		if sdk := detectOsxcrossSDK(osxcrossRoot); sdk != "" {
			extra["macos_sdk"] = sdk
		}
		tools = append(tools, models.Tool{
			Name:  "osxcross",
			Path:  osxcrossRoot,
			Extra: extra,
		})
	}

	// ── mingw-w64 (Windows cross-compile from Linux/macOS) ──────────────────
	mingwTargets := []struct{ name, cmd string }{
		{"mingw-w64 (x86_64)", "x86_64-w64-mingw32-gcc"},
		{"mingw-w64 (i686)", "i686-w64-mingw32-gcc"},
		{"mingw-w64 g++ (x86_64)", "x86_64-w64-mingw32-g++"},
	}
	for _, m := range mingwTargets {
		path, raw := runCmd(m.cmd, "--version")
		if path == "" {
			continue
		}
		version := ""
		parts := strings.Fields(firstLine(raw))
		if len(parts) > 0 {
			version = parts[len(parts)-1]
		}
		tools = append(tools, models.Tool{
			Name:    m.name,
			Version: version,
			Path:    path,
			Extra:   map[string]string{"target": "windows"},
		})
	}

	// ── Wine (run Windows binaries / test cross-compiled output) ────────────
	wineVariants := []struct{ name, cmd string }{
		{"wine", "wine"},
		{"wine64", "wine64"},
	}
	for _, w := range wineVariants {
		path, raw := runCmd(w.cmd, "--version")
		if path == "" {
			continue
		}
		// "wine-8.0.2" or "wine-9.0 (Staging)"
		version := strings.TrimPrefix(firstLine(raw), "wine-")
		version = strings.SplitN(version, " ", 2)[0]

		extra := map[string]string{}
		// Detect Wine prefix
		if prefix := os.Getenv("WINEPREFIX"); prefix != "" {
			extra["wineprefix"] = prefix
		}
		// Check for winetricks
		if _, wt := runCmd("winetricks", "--version"); wt != "" {
			extra["winetricks"] = firstLine(wt)
		}
		// Check for Proton (Steam's Wine fork)
		if proton := detectProton(); proton != "" {
			extra["proton"] = proton
		}

		tools = append(tools, models.Tool{
			Name:    w.name,
			Version: version,
			Path:    path,
			Extra:   extra,
		})
	}

	// ── Zig as a C cross-compiler (zig cc --target=...) ────────────────────
	// Zig ships a built-in C cross-compiler for every target
	if path, raw := runCmd("zig", "version"); path != "" {
		tools = append(tools, models.Tool{
			Name:    "zig-cc",
			Version: firstLine(raw),
			Path:    path,
			Extra: map[string]string{
				"note": "zig cc supports cross-compilation to all major targets",
			},
		})
	}

	// ── Emscripten (compile C/C++ to WebAssembly) ───────────────────────────
	emccVariants := []struct{ name, cmd string }{
		{"emcc", "emcc"},
		{"em++", "em++"},
	}
	for _, e := range emccVariants {
		path, raw := runCmd(e.cmd, "--version")
		if path == "" {
			continue
		}
		// "emcc (Emscripten gcc/clang-like replacement + linker) 3.1.51"
		version := ""
		parts := strings.Fields(firstLine(raw))
		if len(parts) > 0 {
			version = parts[len(parts)-1]
		}
		tools = append(tools, models.Tool{
			Name:    e.name,
			Version: version,
			Path:    path,
			Extra:   map[string]string{"target": "wasm32"},
		})
	}

	// ── Android NDK ──────────────────────────────────────────────────────────
	ndkRoot := os.Getenv("ANDROID_NDK_ROOT")
	if ndkRoot == "" {
		ndkRoot = os.Getenv("ANDROID_NDK_HOME")
	}
	if ndkRoot == "" {
		// Common default NDK paths
		home, _ := os.UserHomeDir()
		candidates := []string{
			filepath.Join(home, "Library", "Android", "sdk", "ndk"),
			filepath.Join(home, "Android", "sdk", "ndk"),
			"/usr/local/android-ndk",
			"/opt/android-ndk",
		}
		for _, c := range candidates {
			// NDK dirs often have a version subdirectory
			if entries, err := os.ReadDir(c); err == nil && len(entries) > 0 {
				ndkRoot = filepath.Join(c, entries[len(entries)-1].Name())
				break
			}
		}
	}
	if ndkRoot != "" {
		version := ""
		// Read version from source.properties
		propFile := filepath.Join(ndkRoot, "source.properties")
		if b, err := os.ReadFile(propFile); err == nil {
			for _, line := range strings.Split(string(b), "\n") {
				if strings.HasPrefix(line, "Pkg.Revision") {
					parts := strings.SplitN(line, "=", 2)
					if len(parts) == 2 {
						version = strings.TrimSpace(parts[1])
					}
				}
			}
		}
		tools = append(tools, models.Tool{
			Name:    "android-ndk",
			Version: version,
			Path:    ndkRoot,
			Extra:   map[string]string{"target": "android"},
		})
	}

	return tools
}

// findDarwinWrappersOnPath scans $PATH for executables matching the pattern
// <arch>-apple-darwin<ver>-clang or similar osxcross wrapper names.
func findDarwinWrappersOnPath() []string {
	var found []string
	seen := map[string]bool{}

	pathDirs := strings.Split(os.Getenv("PATH"), string(os.PathListSeparator))
	for _, dir := range pathDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			name := e.Name()
			if seen[name] {
				continue
			}
			// Match patterns like:
			//   x86_64-apple-darwin23-clang
			//   arm64-apple-darwin22-clang++
			//   aarch64-apple-darwin-clang
			if (strings.Contains(name, "-apple-darwin") && strings.HasSuffix(name, "clang")) ||
				(strings.Contains(name, "-apple-darwin") && strings.HasSuffix(name, "clang++")) ||
				(strings.Contains(name, "-apple-darwin") && strings.HasSuffix(name, "gcc")) {
				found = append(found, name)
				seen[name] = true
			}
		}
	}
	return found
}

// detectOsxcrossSDK tries to find the macOS SDK version inside the osxcross root.
func detectOsxcrossSDK(root string) string {
	if root == "" {
		return ""
	}
	// osxcross stores SDKs in <root>/SDK/MacOSX<ver>.sdk
	sdkDir := filepath.Join(root, "SDK")
	entries, err := os.ReadDir(sdkDir)
	if err != nil {
		// Some builds put them directly in tarballs dir
		sdkDir = filepath.Join(root, "tarballs")
		entries, err = os.ReadDir(sdkDir)
		if err != nil {
			return ""
		}
	}
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, "MacOSX") && (strings.HasSuffix(name, ".sdk") || strings.HasSuffix(name, ".tar.xz") || strings.HasSuffix(name, ".tar.gz")) {
			// Strip prefix and suffix to get version
			ver := strings.TrimPrefix(name, "MacOSX")
			ver = strings.TrimSuffix(ver, ".sdk")
			ver = strings.TrimSuffix(ver, ".tar.xz")
			ver = strings.TrimSuffix(ver, ".tar.gz")
			return ver
		}
	}
	return ""
}

// detectProton looks for Proton installations (Steam's Wine fork).
func detectProton() string {
	home, _ := os.UserHomeDir()
	candidates := []string{
		filepath.Join(home, ".steam", "steam", "steamapps", "common"),
		filepath.Join(home, ".local", "share", "Steam", "steamapps", "common"),
	}
	for _, base := range candidates {
		entries, err := os.ReadDir(base)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), "Proton") {
				return e.Name()
			}
		}
	}
	return ""
}
