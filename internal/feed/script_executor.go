package feed

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
)

// ScriptExecutor handles executing custom scripts for feed fetching
type ScriptExecutor struct {
	scriptsDir string
}

// NewScriptExecutor creates a new ScriptExecutor
func NewScriptExecutor(scriptsDir string) *ScriptExecutor {
	return &ScriptExecutor{scriptsDir: scriptsDir}
}

// findPythonExecutable tries to find a working Python executable
func findPythonExecutable(ctx context.Context) (string, error) {
	// Try different Python executables in order of preference
	candidates := []string{"python", "python3", "py"}

	for _, candidate := range candidates {
		cmd := exec.CommandContext(ctx, candidate, "--version")
		if err := cmd.Run(); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("no Python executable found")
}

// ExecuteScript runs the given script and parses the output as an RSS feed
// The script should output valid RSS/Atom XML to stdout
func (e *ScriptExecutor) ExecuteScript(ctx context.Context, scriptPath string) (*gofeed.Feed, error) {
	// Construct full path
	fullPath := filepath.Join(e.scriptsDir, scriptPath)
	fullPath = filepath.Clean(fullPath)

	// Clean the scripts directory path
	cleanScriptsDir := filepath.Clean(e.scriptsDir)

	// Security check: ensure the script is within the scripts directory
	// Use filepath.Rel to prevent directory traversal attacks
	relPath, err := filepath.Rel(cleanScriptsDir, fullPath)
	if err != nil || strings.HasPrefix(relPath, "..") || strings.Contains(relPath, string(filepath.Separator)+"..") {
		return nil, fmt.Errorf("invalid script path: script must be within scripts directory")
	}

	// Create a context with timeout (30 seconds for script execution)
	execCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Prepare command based on OS and file extension
	var cmd *exec.Cmd
	ext := strings.ToLower(filepath.Ext(fullPath))

	switch ext {
	case ".py":
		// Python script - try to find a working Python executable
		pythonCmd, err := findPythonExecutable(execCtx)
		if err != nil {
			return nil, fmt.Errorf("python script execution failed: %w", err)
		}
		cmd = exec.CommandContext(execCtx, pythonCmd, fullPath)
	case ".sh":
		// Shell script (Unix-like systems)
		if runtime.GOOS == "windows" {
			return nil, fmt.Errorf("shell scripts are not supported on Windows")
		}
		cmd = exec.CommandContext(execCtx, "bash", fullPath)
	case ".ps1":
		// PowerShell script (Windows)
		if runtime.GOOS != "windows" {
			cmd = exec.CommandContext(execCtx, "pwsh", "-File", fullPath)
		} else {
			cmd = exec.CommandContext(execCtx, "powershell.exe", "-ExecutionPolicy", "Bypass", "-File", fullPath)
		}
	case ".js":
		// Node.js script
		cmd = exec.CommandContext(execCtx, "node", fullPath)
	case ".rb":
		// Ruby script
		cmd = exec.CommandContext(execCtx, "ruby", fullPath)
	default:
		// Try to execute directly (for compiled binaries)
		cmd = exec.CommandContext(execCtx, fullPath)
	}

	// Set working directory to the scripts directory
	cmd.Dir = e.scriptsDir

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute the script
	if err := cmd.Run(); err != nil {
		stderrStr := stderr.String()
		if stderrStr != "" {
			return nil, fmt.Errorf("script execution failed: %v, stderr: %s", err, stderrStr)
		}
		return nil, fmt.Errorf("script execution failed: %v", err)
	}

	// Get the script output
	output := stdout.String()

	// Sanitize the XML to remove problematic links (like file:// URLs)
	cleanedOutput := sanitizeFeedXML(output)

	// Parse the sanitized output as RSS/Atom feed
	fp := gofeed.NewParser()
	feed, err := fp.ParseString(cleanedOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to parse script output as feed: %v", err)
	}

	// Fix Atom authors for feeds that use simple text format
	fixFeedAuthors(feed, cleanedOutput)

	return feed, nil
}
