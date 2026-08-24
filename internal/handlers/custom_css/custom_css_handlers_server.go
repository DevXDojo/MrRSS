//go:build server

package custom_css

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
	"MrRSS/internal/utils/fileutil"
)

const customCSSFileName = "custom_article.css"

// HandleUploadCSSDialog is not available in server mode - returns 501 Not Implemented
// @Summary      Upload CSS dialog (not available)
// @Description  File dialog not available in server mode. Use /api/custom-css/upload endpoint instead
// @Tags         custom-css
// @Accept       json
// @Produce      json
// @Success      501  {object}  map[string]string  "Not implemented error"
// @Router       /custom-css/dialog [post]
func HandleUploadCSSDialog(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	log.Printf("File dialog not available in server mode")
	w.WriteHeader(http.StatusNotImplemented)
	response.JSON(w, map[string]interface{}{
		"error": "File dialog not available in server mode. Use /api/custom-css/upload endpoint with file upload instead.",
	})
}

// HandleUploadCSS handles CSS file upload and saves it to the data directory
// @Summary      Upload custom CSS file
// @Description  Upload a custom CSS file to style article content (max 1MB, .css files only)
// @Tags         custom-css
// @Accept       multipart/form-data
// @Produce      json
// @Param        file  formData  file  true  "CSS file to upload"
// @Success      200  {object}  map[string]string  "Upload success (status, message)"
// @Failure      400  {object}  map[string]string  "Bad request (invalid file or size)"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /custom-css/upload [post]
func HandleUploadCSS(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		log.Printf("Error parsing multipart form: %v", err)
		response.Error(w, err, http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		log.Printf("Error getting form file: %v", err)
		response.Error(w, err, http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".css" {
		response.Error(w, fmt.Errorf("only CSS files are allowed"), http.StatusBadRequest)
		return
	}

	// Validate file size (max 1MB)
	if header.Size > 1<<20 {
		response.Error(w, fmt.Errorf("CSS file is too large (max 1MB)"), http.StatusBadRequest)
		return
	}

	// Get data directory
	dataDir, err := fileutil.GetDataDir()
	if err != nil {
		log.Printf("Error getting data directory: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	// Save CSS file
	cssFilePath := filepath.Join(dataDir, customCSSFileName)
	destFile, err := os.Create(cssFilePath)
	if err != nil {
		log.Printf("Error creating CSS file: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}
	defer destFile.Close()

	// Copy file content
	written, err := io.Copy(destFile, file)
	if err != nil {
		log.Printf("Error writing CSS file: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	log.Printf("CSS file uploaded successfully: %s (%d bytes)", header.Filename, written)

	// Update setting in database
	if err := h.DB.SetSetting("custom_css_file", customCSSFileName); err != nil {
		log.Printf("Error saving custom_css_file setting: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	response.JSON(w, map[string]string{
		"status":  "success",
		"message": "CSS file uploaded successfully",
	})
}

// HandleGetCSS returns the custom CSS file content
// @Summary      Get custom CSS
// @Description  Get the content of the uploaded custom CSS file
// @Tags         custom-css
// @Accept       json
// @Produce      text/css
// @Success      200  {string}  string  "CSS file content"
// @Failure      404  {object}  map[string]string  "No custom CSS file configured"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /custom-css [get]
func HandleGetCSS(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	// Get custom_css_file setting
	cssFileName, err := h.DB.GetSetting("custom_css_file")
	if err != nil || cssFileName == "" {
		response.Error(w, fmt.Errorf("no custom CSS file configured"), http.StatusNotFound)
		return
	}

	// Get data directory
	dataDir, err := fileutil.GetDataDir()
	if err != nil {
		log.Printf("Error getting data directory: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	// Read CSS file
	cssFilePath := filepath.Join(dataDir, cssFileName)
	cssContent, err := os.ReadFile(cssFilePath)
	if err != nil {
		log.Printf("Error reading CSS file: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	// Set content type and return CSS
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(cssContent)
}

// HandleDeleteCSS deletes the custom CSS file and clears the setting
// @Summary      Delete custom CSS
// @Description  Delete the custom CSS file and clear the setting
// @Tags         custom-css
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string  "Delete success (status, message)"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /custom-css [delete]
func HandleDeleteCSS(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	// Get custom_css_file setting
	cssFileName, err := h.DB.GetSetting("custom_css_file")
	if err != nil || cssFileName == "" {
		w.WriteHeader(http.StatusOK)
		response.JSON(w, map[string]string{
			"status":  "success",
			"message": "No custom CSS file to delete",
		})
		return
	}

	// Get data directory
	dataDir, err := fileutil.GetDataDir()
	if err != nil {
		log.Printf("Error getting data directory: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	// Delete CSS file
	cssFilePath := filepath.Join(dataDir, cssFileName)
	if err := os.Remove(cssFilePath); err != nil && !os.IsNotExist(err) {
		log.Printf("Error deleting CSS file: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	// Clear setting in database
	if err := h.DB.SetSetting("custom_css_file", ""); err != nil {
		log.Printf("Error clearing custom_css_file setting: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	log.Printf("Custom CSS file deleted: %s", cssFileName)

	// Return success response
	w.WriteHeader(http.StatusOK)
	response.JSON(w, map[string]string{
		"status":  "success",
		"message": "CSS file deleted successfully",
	})
}
