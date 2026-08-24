package update

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
	"MrRSS/internal/utils/fileutil"
)

// HandleInstallUpdate triggers the installation of the downloaded update.
// @Summary      Install update
// @Description  Install the downloaded update (will restart the application)
// @Tags         update
// @Accept       json
// @Produce      json
// @Param        request  body      object  true  "Install request (file_path)"
// @Success      200  {object}  map[string]interface{}  "Installation started (success, message)"
// @Failure      400  {object}  map[string]string  "Bad request (invalid file path or type)"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /update/install [post]
func HandleInstallUpdate(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		FilePath string `json:"file_path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, err, http.StatusBadRequest)
		return
	}

	// Validate file path is within temp directory to prevent path traversal
	tempDir := os.TempDir()
	cleanPath := filepath.Clean(req.FilePath)
	if !strings.HasPrefix(cleanPath, filepath.Clean(tempDir)) {
		log.Printf("Invalid file path attempted: %s", req.FilePath)
		response.Error(w, fmt.Errorf("invalid file path"), http.StatusBadRequest)
		return
	}

	// Validate file exists and is a regular file
	fileInfo, err := os.Stat(cleanPath)
	if os.IsNotExist(err) {
		response.Error(w, fmt.Errorf("update file not found"), http.StatusBadRequest)
		return
	}
	if err != nil {
		log.Printf("Error stating file: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}
	if !fileInfo.Mode().IsRegular() {
		log.Printf("File is not a regular file: %s", cleanPath)
		response.Error(w, fmt.Errorf("invalid file type"), http.StatusBadRequest)
		return
	}

	platform := runtime.GOOS
	isPortable := fileutil.IsPortableMode()
	log.Printf("Installing update from: %s on platform: %s, portable: %v", cleanPath, platform, isPortable)

	// Helper function to schedule cleanup of installer file
	scheduleCleanup := func(filePath string, delay time.Duration) {
		go func() {
			time.Sleep(delay)
			if err := os.Remove(filePath); err != nil {
				log.Printf("Failed to remove installer: %v", err)
			} else {
				log.Printf("Successfully removed installer: %s", filePath)
			}
		}()
	}

	var cmd *exec.Cmd

	if isPortable {
		// Portable mode: extract and replace executable
		if err := installPortableUpdate(cleanPath, platform); err != nil {
			log.Printf("Error installing portable update: %v", err)
			response.Error(w, err, http.StatusInternalServerError)
			return
		}
		scheduleCleanup(cleanPath, 5*time.Second)
		// No need to launch installer, just need to restart the app
	} else {
		// Regular installer mode
		switch platform {
		case "windows":
			// Launch the installer - validate file extension
			if !strings.HasSuffix(strings.ToLower(cleanPath), ".exe") {
				response.Error(w, fmt.Errorf("invalid file type for Windows"), http.StatusBadRequest)
				return
			}
			// Use start command with /B flag to launch in background
			// Format: start /B <executable_path>
			// The /B flag prevents creating a new window
			cmd = exec.Command("cmd.exe", "/C", "start", "/B", cleanPath)
			scheduleCleanup(cleanPath, 10*time.Second)
		case "linux":
			// Make AppImage executable and run it - validate file extension
			if !strings.HasSuffix(strings.ToLower(cleanPath), ".appimage") {
				response.Error(w, fmt.Errorf("invalid file type for Linux"), http.StatusBadRequest)
				return
			}
			if err := os.Chmod(cleanPath, 0755); err != nil {
				log.Printf("Error making file executable: %v", err)
				response.Error(w, fmt.Errorf("failed to prepare installer: %w", err), http.StatusInternalServerError)
				return
			}
			cmd = exec.Command(cleanPath)
			scheduleCleanup(cleanPath, 10*time.Second)
		case "darwin":
			// Open the DMG file - validate file extension
			if !strings.HasSuffix(strings.ToLower(cleanPath), ".dmg") {
				response.Error(w, fmt.Errorf("invalid file type for macOS"), http.StatusBadRequest)
				return
			}
			cmd = exec.Command("open", cleanPath)
			scheduleCleanup(cleanPath, 15*time.Second)
		default:
			response.Error(w, fmt.Errorf("unsupported platform"), http.StatusBadRequest)
			return
		}

		// Start the installer in the background
		if err := cmd.Start(); err != nil {
			log.Printf("Error starting installer: %v", err)
			response.Error(w, fmt.Errorf("failed to start installer: %w", err), http.StatusInternalServerError)
			return
		}

		log.Printf("Installer started successfully, PID: %d", cmd.Process.Pid)
	}

	response.JSON(w, map[string]interface{}{
		"success": true,
		"message": "Installation started. Application will exit shortly.",
	})

	// Schedule graceful shutdown to allow the response to be sent
	// and give time for proper cleanup
	go func() {
		time.Sleep(2 * time.Second)
		log.Println("Initiating graceful shutdown for update installation...")
		// Note: In a production app, this should trigger the Wails shutdown handler
		// which will properly clean up resources. For now, we use os.Exit.
		os.Exit(0)
	}()
}

// installPortableUpdate extracts the portable archive and replaces the current executable
func installPortableUpdate(archivePath string, platform string) error {
	// Get the current executable path and directory
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	exeName := filepath.Base(exePath)

	// Create temporary extraction directory
	extractDir := filepath.Join(os.TempDir(), "mrrss-update-extract")
	if err := os.RemoveAll(extractDir); err != nil {
		log.Printf("Warning: failed to clean extract directory: %v", err)
	}
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return fmt.Errorf("failed to create extract directory: %w", err)
	}
	defer os.RemoveAll(extractDir)

	// Extract archive based on platform
	log.Printf("Extracting archive: %s", archivePath)
	switch platform {
	case "windows", "darwin":
		// .zip format
		if err := extractZip(archivePath, extractDir); err != nil {
			return fmt.Errorf("failed to extract zip: %w", err)
		}
	case "linux":
		// .tar.gz format
		if err := extractTarGz(archivePath, extractDir); err != nil {
			return fmt.Errorf("failed to extract tar.gz: %w", err)
		}
	default:
		return fmt.Errorf("unsupported platform: %s", platform)
	}

	// Find the new executable in the extracted files
	var newExePath string
	err = filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Look for the executable file (MrRSS.exe on Windows, MrRSS on Unix)
		if !info.IsDir() {
			baseName := filepath.Base(path)
			if platform == "windows" && strings.EqualFold(baseName, exeName) {
				newExePath = path
				return filepath.SkipDir
			} else if platform != "windows" && baseName == exeName {
				newExePath = path
				return filepath.SkipDir
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to search for executable: %w", err)
	}
	if newExePath == "" {
		return fmt.Errorf("executable not found in archive")
	}

	log.Printf("Found new executable: %s", newExePath)

	// Create backup of current executable
	backupPath := exePath + ".backup"
	if err := os.Rename(exePath, backupPath); err != nil {
		return fmt.Errorf("failed to backup current executable: %w", err)
	}

	// Copy new executable to replace the old one
	if err := copyFile(newExePath, exePath); err != nil {
		// Restore backup if copy fails
		if restoreErr := os.Rename(backupPath, exePath); restoreErr != nil {
			log.Printf("CRITICAL: Failed to restore backup: %v", restoreErr)
		}
		return fmt.Errorf("failed to copy new executable: %w", err)
	}

	// Set executable permissions on Unix systems
	if platform != "windows" {
		if err := os.Chmod(exePath, 0755); err != nil {
			log.Printf("Warning: failed to set executable permissions: %v", err)
		}
	}

	// Schedule backup cleanup
	go func() {
		time.Sleep(10 * time.Second)
		if err := os.Remove(backupPath); err != nil {
			log.Printf("Failed to remove backup file: %v", err)
		} else {
			log.Printf("Successfully removed backup file: %s", backupPath)
		}
	}()

	log.Printf("Portable update installed successfully")
	return nil
}

// extractZip extracts a ZIP archive to the specified directory
func extractZip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		// Prevent path traversal
		fpath := filepath.Join(destDir, f.Name)
		if !strings.HasPrefix(filepath.Clean(fpath), filepath.Clean(destDir)) {
			return fmt.Errorf("invalid file path in archive: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, 0755)
			continue
		}

		// Create parent directory if needed
		if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
			return err
		}

		// Extract file
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

// extractTarGz extracts a .tar.gz archive to the specified directory
func extractTarGz(tarGzPath, destDir string) error {
	file, err := os.Open(tarGzPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Prevent path traversal
		fpath := filepath.Join(destDir, header.Name)
		if !strings.HasPrefix(filepath.Clean(fpath), filepath.Clean(destDir)) {
			return fmt.Errorf("invalid file path in archive: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(fpath, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			// Create parent directory if needed
			if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
				return err
			}

			// Extract file
			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}
	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	return destFile.Sync()
}
