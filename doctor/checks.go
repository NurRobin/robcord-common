package doctor

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

// CheckDataDir verifies that the data directory exists and is writable.
// Auto-fix: creates the directory with mode 0755.
func CheckDataDir(dataDir string) CheckFunc {
	return func(autoFix bool) CheckResult {
		info, err := os.Stat(dataDir)
		if err != nil {
			if !os.IsNotExist(err) {
				return CheckResult{
					Severity: Fatal,
					Message:  fmt.Sprintf("cannot stat data dir: %v", err),
					FixHint:  fmt.Sprintf("mkdir -p '%s' && chmod 0755 '%s'", dataDir, dataDir),
				}
			}
			if autoFix {
				if mkErr := os.MkdirAll(dataDir, 0755); mkErr != nil {
					return CheckResult{
						Severity: Fatal,
						Message:  fmt.Sprintf("auto-fix failed: %v", mkErr),
					}
				}
				return CheckResult{Passed: true, Fixed: true, Message: "created data directory"}
			}
			return CheckResult{
				Severity: Fatal,
				Message:  fmt.Sprintf("data directory does not exist: %s", dataDir),
				FixHint:  fmt.Sprintf("mkdir -p '%s' && chmod 0755 '%s'", dataDir, dataDir),
			}
		}
		if !info.IsDir() {
			return CheckResult{
				Severity: Fatal,
				Message:  fmt.Sprintf("%s exists but is not a directory", dataDir),
			}
		}
		// Verify the directory is actually writable
		tmp := filepath.Join(dataDir, ".doctor_dir_test")
		defer os.Remove(tmp)
		f, err := os.Create(tmp)
		if err != nil {
			return CheckResult{
				Severity: Fatal,
				Message:  fmt.Sprintf("data directory exists but is not writable: %v", err),
				FixHint:  fmt.Sprintf("check ownership and permissions on '%s'", dataDir),
			}
		}
		f.Close()
		return CheckResult{Passed: true}
	}
}

// CheckDBFile verifies that the database file exists and has correct permissions.
// Auto-fix: chmod 0600.
func CheckDBFile(dataDir, filename string) CheckFunc {
	return func(autoFix bool) CheckResult {
		path := filepath.Join(dataDir, filename)
		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				return CheckResult{
					Passed:  true,
					Message: "database file does not exist yet (will be created on first run)",
				}
			}
			return CheckResult{
				Severity: Warning,
				Message:  fmt.Sprintf("cannot stat database file: %v", err),
			}
		}
		if runtime.GOOS != "windows" {
			mode := info.Mode().Perm()
			if mode&0o077 != 0 {
				if autoFix {
					if chErr := os.Chmod(path, 0600); chErr != nil {
						return CheckResult{
							Severity: Warning,
							Message:  fmt.Sprintf("auto-fix chmod failed: %v", chErr),
						}
					}
					return CheckResult{Passed: true, Fixed: true, Message: "fixed database file permissions to 0600"}
				}
				return CheckResult{
					Severity: Warning,
					Message:  fmt.Sprintf("database file has group/other permissions (mode %04o)", mode),
					FixHint:  fmt.Sprintf("chmod 0600 '%s'", path),
				}
			}
		}
		return CheckResult{Passed: true}
	}
}

// CheckEnvFilePerms warns if a local .env file has overly permissive mode.
// Auto-fix: chmod 0600.
func CheckEnvFilePerms(path string) CheckFunc {
	return func(autoFix bool) CheckResult {
		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				return CheckResult{Passed: true, Message: "no .env file in current directory"}
			}
			return CheckResult{
				Severity: Warning,
				Message:  fmt.Sprintf("cannot stat .env file: %v", err),
			}
		}
		if runtime.GOOS == "windows" {
			return CheckResult{Passed: true}
		}
		mode := info.Mode().Perm()
		if mode&0o077 != 0 {
			if autoFix {
				if chErr := os.Chmod(path, 0600); chErr != nil {
					return CheckResult{
						Severity: Warning,
						Message:  fmt.Sprintf("auto-fix chmod failed: %v", chErr),
					}
				}
				return CheckResult{Passed: true, Fixed: true, Message: "fixed .env permissions to 0600"}
			}
			return CheckResult{
				Severity: Warning,
				Message:  fmt.Sprintf(".env has permissive mode %04o (should be 0600)", mode),
				FixHint:  fmt.Sprintf("chmod 0600 '%s'", path),
			}
		}
		return CheckResult{Passed: true}
	}
}

// CheckPortAvailable verifies that a TCP port is not already in use.
// portEnvVar is the environment variable name shown in the fix hint (e.g.
// "ZENTRALE_API_PORT" or "WORKSPACE_PORT").
// Note: this is a best-effort check with an inherent TOCTOU race — another
// process could bind the port between this check and the actual server start.
func CheckPortAvailable(port int, portEnvVar string) CheckFunc {
	return func(autoFix bool) CheckResult {
		ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
		if err != nil {
			return CheckResult{
				Severity: Fatal,
				Message:  fmt.Sprintf("cannot listen on TCP port %d: %v", port, err),
				FixHint:  fmt.Sprintf("stop the process using port %d, or set %s to a different port", port, portEnvVar),
			}
		}
		ln.Close()
		return CheckResult{Passed: true}
	}
}

// CheckDockerVolume warns if running in Docker and the data directory is not
// writable by the container user.
func CheckDockerVolume(dataDir string) CheckFunc {
	return func(autoFix bool) CheckResult {
		if _, err := os.Stat("/.dockerenv"); os.IsNotExist(err) {
			return CheckResult{Passed: true, Message: "not running in Docker"}
		}
		info, err := os.Stat(dataDir)
		if err != nil {
			return CheckResult{
				Severity: Warning,
				Message:  fmt.Sprintf("cannot stat data dir in Docker: %v", err),
			}
		}
		if !info.IsDir() {
			return CheckResult{
				Severity: Warning,
				Message:  "data path is not a directory",
			}
		}
		tmp := filepath.Join(dataDir, ".docker_volume_test")
		defer os.Remove(tmp)
		f, err := os.Create(tmp)
		if err != nil {
			return CheckResult{
				Severity: Warning,
				Message:  "Docker volume is not writable — check that the data volume has correct ownership",
				FixHint:  fmt.Sprintf("chown -R $(id -u):$(id -g) %s", dataDir),
			}
		}
		f.Close()
		return CheckResult{Passed: true, Message: "Docker volume is writable"}
	}
}
