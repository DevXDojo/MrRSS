package utils

import (
	"fmt"
	"html"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"MrRSS/internal/utils/fileutil"
)

// EnableStartup enables the application to start on system boot
func EnableStartup() error {
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	switch runtime.GOOS {
	case "windows":
		return enableStartupWindows(executable)
	case "linux":
		return enableStartupLinux(executable)
	case "darwin":
		return enableStartupDarwin(executable)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// DisableStartup disables the application from starting on system boot
func DisableStartup() error {
	switch runtime.GOOS {
	case "windows":
		return disableStartupWindows()
	case "linux":
		return disableStartupLinux()
	case "darwin":
		return disableStartupDarwin()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// Windows implementation using registry
func enableStartupWindows(executable string) error {
	command := `"` + executable + `"`
	if dir := fileutil.CustomDataDir(); dir != "" {
		// Windows paths cannot contain quotes; double trailing slashes before
		// the closing quote (notably when the chosen directory is a drive root).
		trimmed := strings.TrimRight(dir, `\`)
		command += ` --data-dir "` + trimmed + strings.Repeat(`\`, 2*(len(dir)-len(trimmed))) + `"`
	}
	// Use reg.exe to add registry entry
	cmd := exec.Command("reg", "add",
		"HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Run",
		"/v", "MrRSS",
		"/t", "REG_SZ",
		"/d", command,
		"/f")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add registry entry: %v, output: %s", err, output)
	}

	log.Printf("Startup enabled for Windows: %s", executable)
	return nil
}

func disableStartupWindows() error {
	// Use reg.exe to remove registry entry
	cmd := exec.Command("reg", "delete",
		"HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Run",
		"/v", "MrRSS",
		"/f")

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if the error is because the key doesn't exist (exit code 1)
		// If so, we can ignore it since our goal is to have the key not present
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			log.Println("Startup registry key was not present (already disabled)")
			return nil
		}
		return fmt.Errorf("failed to remove registry entry: %v, output: %s", err, output)
	}

	log.Println("Startup disabled for Windows")
	return nil
}

// Linux implementation using .desktop file in autostart
func enableStartupLinux(executable string) error {
	// APPIMAGE identifies the persistent outer file, not the transient mount.
	if appImage := os.Getenv("APPIMAGE"); appImage != "" {
		executable = appImage
	}
	execValue, err := desktopExec(executable)
	if err != nil {
		return err
	}
	if dir := fileutil.CustomDataDir(); dir != "" {
		quotedDir, err := desktopExec(dir)
		if err != nil {
			return err
		}
		execValue += " --data-dir " + quotedDir
	}
	autostartDir, err := linuxAutostartDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(autostartDir, 0755); err != nil {
		return fmt.Errorf("failed to create autostart directory: %w", err)
	}

	desktopFile := filepath.Join(autostartDir, "mrrss.desktop")
	content := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=MrRSS
Exec=%s
Hidden=false
NoDisplay=false
X-GNOME-Autostart-enabled=true
`, execValue)

	if err := os.WriteFile(desktopFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write desktop file: %w", err)
	}

	log.Printf("Startup enabled for Linux: %s", desktopFile)
	return nil
}

func disableStartupLinux() error {
	autostartDir, err := linuxAutostartDir()
	if err != nil {
		return err
	}

	desktopFile := filepath.Join(autostartDir, "mrrss.desktop")
	if err := os.Remove(desktopFile); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove desktop file: %w", err)
		}
	}

	log.Println("Startup disabled for Linux")
	return nil
}

func linuxAutostartDir() (string, error) {
	if configDir := os.Getenv("XDG_CONFIG_HOME"); filepath.IsAbs(configDir) {
		return filepath.Join(configDir, "autostart"), nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".config", "autostart"), nil
}

// desktopExec quotes one executable according to the Desktop Entry Exec grammar.
// Escape argument quoting first, then the desktop file's string escaping layer.
func desktopExec(executable string) (string, error) {
	if !filepath.IsAbs(executable) || strings.ContainsAny(executable, "\x00\r\n") {
		return "", fmt.Errorf("startup executable must be an absolute path without line breaks")
	}
	escaped := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "`", "\\`", "$", `\$`, "%", "%%").Replace(executable)
	escaped = strings.NewReplacer(`\`, `\\`, "\t", `\t`).Replace(escaped)
	return `"` + escaped + `"`, nil
}

// macOS implementation using LaunchAgents plist
func enableStartupDarwin(executable string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	launchAgentsDir := filepath.Join(homeDir, "Library", "LaunchAgents")
	if err := os.MkdirAll(launchAgentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create LaunchAgents directory: %w", err)
	}

	plistFile := filepath.Join(launchAgentsDir, "com.mrrss.app.plist")
	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.mrrss.app</string>
	<key>ProgramArguments</key>
	<array>
		%s
	</array>
	<key>RunAtLoad</key>
	<true/>
</dict>
</plist>
`, startupDarwinArguments(executable))

	if err := os.WriteFile(plistFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write plist file: %w", err)
	}

	log.Printf("Startup enabled for macOS: %s", plistFile)
	return nil
}

func startupDarwinArguments(executable string) string {
	args := "<string>" + html.EscapeString(executable) + "</string>"
	if dir := fileutil.CustomDataDir(); dir != "" {
		args += "\n\t\t<string>--data-dir</string>\n\t\t<string>" + html.EscapeString(dir) + "</string>"
	}
	return args
}

func disableStartupDarwin() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	plistFile := filepath.Join(homeDir, "Library", "LaunchAgents", "com.mrrss.app.plist")
	if err := os.Remove(plistFile); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove plist file: %w", err)
		}
	}

	log.Println("Startup disabled for macOS")
	return nil
}
