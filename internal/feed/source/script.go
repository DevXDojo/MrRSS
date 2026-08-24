package source

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
)

// ScriptSource executes custom scripts to fetch feed content.
type ScriptSource struct {
	scriptsDir string
}

// NewScriptSource creates a new script source with the given scripts directory.
func NewScriptSource(scriptsDir string) *ScriptSource {
	return &ScriptSource{scriptsDir: scriptsDir}
}

// Type returns the source type identifier.
func (s *ScriptSource) Type() Type {
	return TypeScript
}

// Validate checks if the configuration is valid for script source.
func (s *ScriptSource) Validate(config *Config) error {
	if config == nil {
		return errors.New("config is nil")
	}
	if config.ScriptPath == "" {
		return errors.New("script path is required for script source")
	}
	if s.scriptsDir == "" {
		return errors.New("scripts directory is not configured")
	}

	// Validate path to prevent directory traversal
	fullPath := filepath.Join(s.scriptsDir, config.ScriptPath)
	fullPath = filepath.Clean(fullPath)
	cleanScriptsDir := filepath.Clean(s.scriptsDir)

	relPath, err := filepath.Rel(cleanScriptsDir, fullPath)
	if err != nil || strings.HasPrefix(relPath, "..") {
		return errors.New("invalid script path: must be within scripts directory")
	}

	return nil
}

// Fetch executes the script and parses its output as RSS feed.
func (s *ScriptSource) Fetch(ctx context.Context, config *Config) (*gofeed.Feed, error) {
	if err := s.Validate(config); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	fullPath := filepath.Join(s.scriptsDir, config.ScriptPath)
	fullPath = filepath.Clean(fullPath)

	// Set timeout
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Prepare command based on file extension
	cmd, err := s.buildCommand(execCtx, fullPath)
	if err != nil {
		return nil, err
	}

	// Set working directory
	cmd.Dir = s.scriptsDir

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute
	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return nil, fmt.Errorf("script failed: %v, stderr: %s", err, stderr.String())
		}
		return nil, fmt.Errorf("script failed: %v", err)
	}

	// Parse output as RSS
	parser := gofeed.NewParser()
	feed, err := parser.ParseString(stdout.String())
	if err != nil {
		return nil, fmt.Errorf("failed to parse script output: %w", err)
	}

	return feed, nil
}

// buildCommand creates the exec.Cmd for the given script.
func (s *ScriptSource) buildCommand(ctx context.Context, fullPath string) (*exec.Cmd, error) {
	ext := strings.ToLower(filepath.Ext(fullPath))

	switch ext {
	case ".py":
		pythonCmd, err := findPython(ctx)
		if err != nil {
			return nil, err
		}
		return exec.CommandContext(ctx, pythonCmd, fullPath), nil

	case ".sh":
		if runtime.GOOS == "windows" {
			return nil, errors.New("shell scripts are not supported on Windows")
		}
		return exec.CommandContext(ctx, "bash", fullPath), nil

	case ".ps1":
		if runtime.GOOS != "windows" {
			return nil, errors.New("PowerShell scripts are only supported on Windows")
		}
		return exec.CommandContext(ctx, "powershell", "-ExecutionPolicy", "Bypass", "-File", fullPath), nil

	case ".js":
		return exec.CommandContext(ctx, "node", fullPath), nil

	case ".rb":
		return exec.CommandContext(ctx, "ruby", fullPath), nil

	default:
		// Try to execute directly
		return exec.CommandContext(ctx, fullPath), nil
	}
}

// findPython tries to find a working Python executable.
func findPython(ctx context.Context) (string, error) {
	candidates := []string{"python", "python3", "py"}

	for _, candidate := range candidates {
		cmd := exec.CommandContext(ctx, candidate, "--version")
		if err := cmd.Run(); err == nil {
			return candidate, nil
		}
	}

	return "", errors.New("no Python executable found")
}
