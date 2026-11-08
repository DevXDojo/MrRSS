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
// @Description  Desktop-only installation of an update verified in this process; requires a native capability and rechecks SHA-256. Disabled in server mode.
// @Tags         update
// @Param        Authorization header string true "Native desktop capability (Bearer token)"
// @Failure      403  {object}  map[string]string  "Desktop update authorization required"
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
	if !authorizeUpdate(h, w, r) {
		return
	}

	var req struct {
		FilePath string `json:"file_path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, err, http.StatusBadRequest)
		return
	}

	// Accept only a regular file in a dedicated update download directory.
	cleanPath, err := validateUpdateDownload(req.FilePath)
	if err != nil {
		log.Printf("Invalid file path attempted: %s", req.FilePath)
		response.Error(w, fmt.Errorf("invalid file path"), http.StatusBadRequest)
		return
	}

	digest, ok := verifiedUpdates.LoadAndDelete(verifiedUpdateKey{h, cleanPath})
	if !ok || verifyUpdateDigest(r.Context(), cleanPath, fmt.Sprint(digest)) != nil {
		response.Error(w, fmt.Errorf("update must be downloaded and verified in this session"), http.StatusBadRequest)
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

	// AppImages run from a read-only mount even when portable.txt is bundled.
	// Always update the persistent outer image before considering portable mode.
	if platform == "linux" && os.Getenv("APPIMAGE") != "" {
		target, err := replaceAppImage(r.Context(), cleanPath, os.Getenv("APPIMAGE"))
		if err != nil {
			log.Printf("AppImage update failed: %v", err)
			response.Error(w, fmt.Errorf("cannot replace installed AppImage; check directory permissions or update it manually"), http.StatusInternalServerError)
			return
		}
		if err := startAppImageAfterExit(target, fileutil.DataDirRestartArgs(os.Args[1:])); err != nil {
			log.Printf("AppImage restart failed: %v", err)
			response.Error(w, fmt.Errorf("update installed; quit and reopen the installed AppImage manually"), http.StatusInternalServerError)
			return
		}
		cleanupUpdateDownload(cleanPath)
		response.JSON(w, map[string]interface{}{"success": true, "message": "Update installed. Application will restart shortly."})
		go func() {
			time.Sleep(2 * time.Second)
			if h.QuitForUpdate != nil {
				h.QuitForUpdate()
			} else {
				os.Exit(0)
			}
		}()
		return
	}

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
		cleanupUpdateDownload(cleanPath)
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
			// Execute the verified installer directly; never interpret its path as shell code.
			cmd = exec.Command(cleanPath)
			scheduleCleanup(cleanPath, 10*time.Second)
		case "linux":
			response.Error(w, fmt.Errorf("installed AppImage path is unavailable; please install the update manually"), http.StatusBadRequest)
			return
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
	extractDir, err := os.MkdirTemp("", "mrrss-update-extract-")
	if err != nil {
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

// archivePath validates names consistently on every operating system.
func archivePath(name string) (string, error) {
	if strings.ContainsAny(name, "\\:") || !filepath.IsLocal(name) {
		return "", fmt.Errorf("invalid archive path")
	}
	for _, part := range strings.Split(name, "/") {
		if part == ".." {
			return "", fmt.Errorf("invalid archive path")
		}
	}
	return filepath.FromSlash(name), nil
}

// Extraction uses an OS-enforced root, including when an existing path is a
// symlink. Only directories and ordinary files are accepted from archives.
func extractZip(zipPath, destDir string) error {
	archive, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer archive.Close()
	root, err := os.OpenRoot(destDir)
	if err != nil {
		return err
	}
	defer root.Close()
	for _, file := range archive.File {
		name, err := archivePath(file.Name)
		if err != nil {
			return err
		}
		if file.FileInfo().IsDir() {
			if err := root.MkdirAll(name, 0755); err != nil {
				return err
			}
			continue
		}
		if !file.Mode().IsRegular() {
			return fmt.Errorf("unsupported archive entry")
		}
		if err := root.MkdirAll(filepath.Dir(name), 0755); err != nil {
			return err
		}
		source, err := file.Open()
		if err != nil {
			return err
		}
		err = extractRegularFile(root, name, file.Mode().Perm(), source)
		closeErr := source.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}

func extractTarGz(tarGzPath, destDir string) error {
	file, err := os.Open(tarGzPath)
	if err != nil {
		return err
	}
	defer file.Close()
	gz, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gz.Close()
	root, err := os.OpenRoot(destDir)
	if err != nil {
		return err
	}
	defer root.Close()
	reader := tar.NewReader(gz)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		name, err := archivePath(header.Name)
		if err != nil {
			return err
		}
		switch header.Typeflag {
		case tar.TypeDir:
			if err := root.MkdirAll(name, 0755); err != nil {
				return err
			}
		case tar.TypeReg, tar.TypeRegA:
			if err := root.MkdirAll(filepath.Dir(name), 0755); err != nil {
				return err
			}
			if err := extractRegularFile(root, name, os.FileMode(header.Mode).Perm(), reader); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported archive entry")
		}
	}
}

func extractRegularFile(root *os.Root, name string, mode os.FileMode, source io.Reader) error {
	file, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode.Perm())
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(file, source)
	closeErr := file.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
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
