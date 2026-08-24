//go:build !server

package opml

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
	"MrRSS/internal/jsonimport"
	"MrRSS/internal/models"
	"MrRSS/internal/opml"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// saveFeedTags creates or finds tags by name and associates them with the feed
func saveFeedTags(h *core.Handler, feedID int64, tags []models.Tag) error {
	if len(tags) == 0 {
		return nil
	}

	var tagIDs []int64
	for _, tag := range tags {
		// Check if tag already exists by name
		existingTags, err := h.DB.GetTags()
		if err != nil {
			log.Printf("Error fetching tags: %v", err)
			continue
		}

		var foundTagID int64
		for _, existingTag := range existingTags {
			if strings.EqualFold(existingTag.Name, tag.Name) {
				foundTagID = existingTag.ID
				break
			}
		}

		// If tag doesn't exist, create it
		if foundTagID == 0 {
			newTag := &models.Tag{
				Name:  tag.Name,
				Color: tag.Color,
			}
			id, err := h.DB.AddTag(newTag)
			if err != nil {
				log.Printf("Error creating tag %s: %v", tag.Name, err)
				continue
			}
			foundTagID = id
		}

		tagIDs = append(tagIDs, foundTagID)
	}

	// Associate tags with feed
	if len(tagIDs) > 0 {
		return h.DB.SetFeedTags(feedID, tagIDs)
	}

	return nil
}

// HandleOPMLImport handles OPML/JSON file import based on file extension.
// @Summary      Import subscriptions from OPML/JSON
// @Description  Import RSS feed subscriptions from an OPML or JSON file
// @Tags         opml
// @Accept       multipart/form-data
// @Produce      json
// @Param        file  formData  file  false  "OPML or JSON file to import"
// @Success      200  {object}  map[string]string  "Import successful"
// @Failure      400  {object}  map[string]string  "Bad request"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /opml/import [post]
func HandleOPMLImport(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleOPMLImport: ContentLength: %d", r.ContentLength)
	contentType := r.Header.Get("Content-Type")
	log.Printf("HandleOPMLImport: Content-Type: %s", contentType)

	var file io.Reader
	var filename string

	if strings.Contains(contentType, "multipart/form-data") {
		f, header, err := r.FormFile("file")
		if err != nil {
			log.Printf("Error getting form file: %v", err)
			response.Error(w, err, http.StatusBadRequest)
			return
		}
		defer f.Close()
		filename = header.Filename
		log.Printf("HandleOPMLImport: Received file %s, size: %d", filename, header.Size)

		if header.Size == 0 {
			response.Error(w, nil, http.StatusBadRequest)
			return
		}
		file = f
	} else {
		// Handle raw body upload
		file = r.Body
		defer r.Body.Close()
	}

	// Determine format based on file extension
	ext := strings.ToLower(filepath.Ext(filename))
	isJSON := ext == ".json"

	var feeds []models.Feed
	var err error

	if isJSON {
		log.Printf("HandleOPMLImport: Detected JSON format from extension %s", ext)
		feeds, err = jsonimport.Parse(file)
	} else {
		log.Printf("HandleOPMLImport: Using OPML format (extension: %s)", ext)
		feeds, err = opml.Parse(file)
	}

	if err != nil {
		log.Printf("Error parsing file: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	// Import feeds synchronously so they appear in the sidebar immediately
	var feedIDs []int64
	for _, f := range feeds {
		var feedID int64
		var err error

		// Check if feed has XPath configuration
		if f.Type == "HTML+XPath" || f.Type == "XML+XPath" {
			feedID, err = h.Fetcher.AddXPathSubscription(
				f.URL, f.Category, f.Title, f.Type,
				f.XPathItem, f.XPathItemTitle, f.XPathItemContent, f.XPathItemUri,
				f.XPathItemAuthor, f.XPathItemTimestamp, f.XPathItemTimeFormat,
				f.XPathItemThumbnail, f.XPathItemCategories, f.XPathItemUid,
			)
		} else {
			feedID, err = h.Fetcher.ImportSubscription(f.Title, f.URL, f.Category)
		}

		if err != nil {
			log.Printf("Error importing feed %s: %v", f.Title, err)
			continue
		}

		// Save tags for the feed
		if len(f.Tags) > 0 {
			if err := saveFeedTags(h, feedID, f.Tags); err != nil {
				log.Printf("Error saving tags for feed %s: %v", f.Title, err)
				// Continue even if tag saving fails
			}
		}

		feedIDs = append(feedIDs, feedID)
	}

	// Fetch articles for the newly imported feeds asynchronously with progress tracking
	if len(feedIDs) > 0 {
		go func() {
			h.Fetcher.FetchFeedsByIDs(context.Background(), feedIDs)
		}()
	}

	w.WriteHeader(http.StatusOK)
}

// HandleOPMLExport handles OPML file export.
// @Summary      Export subscriptions to OPML
// @Description  Export all local RSS feed subscriptions to an OPML file (excludes FreshRSS feeds)
// @Tags         opml
// @Accept       json
// @Produce      text/xml
// @Success      200  {string}  string  "OPML file content"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /opml/export [get]
func HandleOPMLExport(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	feeds, err := h.DB.GetFeeds()
	if err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	// Filter out FreshRSS feeds - only export local feeds
	localFeeds := make([]models.Feed, 0)
	for _, feed := range feeds {
		if !feed.IsFreshRSSSource {
			localFeeds = append(localFeeds, feed)
		}
	}

	log.Printf("[OPML Export] Exporting %d local feeds (excluded %d FreshRSS feeds)",
		len(localFeeds), len(feeds)-len(localFeeds))

	data, err := opml.Generate(localFeeds)
	if err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=subscriptions.opml")
	w.Header().Set("Content-Type", "text/xml")
	w.Write(data)
}

// HandleOPMLImportDialog opens a file dialog to select OPML file for import.
// @Summary      Import dialog (desktop mode)
// @Description  Open a file dialog to select an OPML or JSON file for import (desktop mode only)
// @Tags         opml
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "Import success (status, feedCount, filePath)"
// @Success      501  {object}  map[string]string  "Not implemented in server mode"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /opml/import/dialog [post]
func HandleOPMLImportDialog(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if h.App == nil {
		log.Printf("File dialog not available")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotImplemented)
		response.JSON(w, map[string]interface{}{
			"error": "File dialog not available. Use /api/opml/import endpoint with file upload instead.",
		})
		return
	}

	// Type assert to *application.App to access Dialog
	app, ok := h.App.(*application.App)
	if !ok {
		log.Printf("File dialog not available: app is not *application.App type")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotImplemented)
		response.JSON(w, map[string]interface{}{
			"error": "File dialog not available. Use /api/opml/import endpoint with file upload instead.",
		})
		return
	}

	filePath, err := app.Dialog.OpenFileWithOptions(&application.OpenFileDialogOptions{
		Title: "Import Subscriptions",
		Filters: []application.FileFilter{
			{
				DisplayName: "Supported Files (*.opml;*.xml;*.json)",
				Pattern:     "*.opml;*.xml;*.json",
			},
			{
				DisplayName: "OPML Files (*.opml;*.xml)",
				Pattern:     "*.opml;*.xml",
			},
			{
				DisplayName: "JSON Files (*.json)",
				Pattern:     "*.json",
			},
			{
				DisplayName: "All Files (*)",
				Pattern:     "*",
			},
		},
		CanChooseFiles:       true,
		AllowsOtherFileTypes: true,
	}).PromptForSingleSelection()

	// Treat empty filePath as user cancellation (no error should be shown)
	if filePath == "" {
		log.Printf("Import dialog cancelled by user")
		w.Header().Set("Content-Type", "application/json")
		response.JSON(w, map[string]string{"status": "cancelled"})
		return
	}

	// Only show error for actual failures, not cancellations
	if err != nil {
		log.Printf("Error opening file dialog: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		response.JSON(w, map[string]interface{}{
			"error": "Failed to open file dialog",
		})
		return
	}

	// Read the selected file
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Error opening selected file: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		response.JSON(w, map[string]interface{}{
			"error": "Failed to open selected file",
		})
		return
	}
	defer file.Close()

	// Determine format based on file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	isJSON := ext == ".json"

	var feeds []models.Feed

	if isJSON {
		log.Printf("HandleOPMLImportDialog: Detected JSON format from extension %s", ext)
		feeds, err = jsonimport.Parse(file)
	} else {
		log.Printf("HandleOPMLImportDialog: Using OPML format (extension: %s)", ext)
		feeds, err = opml.Parse(file)
	}

	if err != nil {
		log.Printf("Error parsing file: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		response.JSON(w, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Import feeds synchronously so they appear in the sidebar immediately
	var feedIDs []int64
	for _, f := range feeds {
		var feedID int64
		var err error

		// Check if feed has XPath configuration
		if f.Type == "HTML+XPath" || f.Type == "XML+XPath" {
			feedID, err = h.Fetcher.AddXPathSubscription(
				f.URL, f.Category, f.Title, f.Type,
				f.XPathItem, f.XPathItemTitle, f.XPathItemContent, f.XPathItemUri,
				f.XPathItemAuthor, f.XPathItemTimestamp, f.XPathItemTimeFormat,
				f.XPathItemThumbnail, f.XPathItemCategories, f.XPathItemUid,
			)
		} else {
			feedID, err = h.Fetcher.ImportSubscription(f.Title, f.URL, f.Category)
		}

		if err != nil {
			log.Printf("Error importing feed %s: %v", f.Title, err)
			continue
		}

		// Save tags for the feed
		if len(f.Tags) > 0 {
			if err := saveFeedTags(h, feedID, f.Tags); err != nil {
				log.Printf("Error saving tags for feed %s: %v", f.Title, err)
				// Continue even if tag saving fails
			}
		}

		feedIDs = append(feedIDs, feedID)
	}

	// Fetch articles for the newly imported feeds asynchronously with progress tracking
	if len(feedIDs) > 0 {
		go func() {
			h.Fetcher.FetchFeedsByIDs(context.Background(), feedIDs)
		}()
	}

	w.Header().Set("Content-Type", "application/json")
	response.JSON(w, map[string]interface{}{
		"status":    "success",
		"feedCount": len(feeds),
		"filePath":  filePath,
	})
}

// HandleOPMLExportDialog opens a save dialog to export OPML file.
// @Summary      Export dialog (desktop mode)
// @Description  Open a save dialog to export subscriptions to OPML or JSON file (desktop mode only)
// @Tags         opml
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "Export success (status, filePath)"
// @Success      501  {object}  map[string]string  "Not implemented in server mode"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /opml/export/dialog [post]
func HandleOPMLExportDialog(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if h.App == nil {
		log.Printf("File dialog not available")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotImplemented)
		response.JSON(w, map[string]interface{}{
			"error": "File dialog not available. Use the direct export endpoint instead.",
		})
		return
	}

	// Get feeds data
	feeds, err := h.DB.GetFeeds()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		response.JSON(w, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Filter out FreshRSS feeds - only export local feeds
	localFeeds := make([]models.Feed, 0)
	for _, feed := range feeds {
		if !feed.IsFreshRSSSource {
			localFeeds = append(localFeeds, feed)
		}
	}

	log.Printf("[OPML Export Dialog] Exporting %d local feeds (excluded %d FreshRSS feeds)",
		len(localFeeds), len(feeds)-len(localFeeds))

	// Type assert to *application.App to access Dialog
	app, ok := h.App.(*application.App)
	if !ok {
		log.Printf("File dialog not available: app is not *application.App type")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotImplemented)
		response.JSON(w, map[string]interface{}{
			"error": "File dialog not available. Use /api/opml/export endpoint with direct download instead.",
		})
		return
	}

	filePath, err := app.Dialog.SaveFileWithOptions(&application.SaveFileDialogOptions{
		Title:    "Export Subscriptions",
		Filename: "subscriptions.opml",
		Filters: []application.FileFilter{
			{
				DisplayName: "OPML Files (*.opml)",
				Pattern:     "*.opml",
			},
			{
				DisplayName: "JSON Files (*.json)",
				Pattern:     "*.json",
			},
			{
				DisplayName: "XML Files (*.xml)",
				Pattern:     "*.xml",
			},
			{
				DisplayName: "All Files (*)",
				Pattern:     "*",
			},
		},
		AllowOtherFileTypes: true,
	}).PromptForSingleSelection()

	// Treat empty filePath as user cancellation (no error should be shown)
	if filePath == "" {
		log.Printf("Export dialog cancelled by user")
		w.Header().Set("Content-Type", "application/json")
		response.JSON(w, map[string]string{"status": "cancelled"})
		return
	}

	// Only show error for actual failures, not cancellations
	if err != nil {
		log.Printf("Error opening save dialog: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		response.JSON(w, map[string]interface{}{
			"error": "Failed to open save dialog",
		})
		return
	}

	// Determine format based on file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	isJSON := ext == ".json"

	var data []byte
	if isJSON {
		log.Printf("HandleOPMLExportDialog: Generating JSON format")
		data, err = jsonimport.Generate(localFeeds)
	} else {
		log.Printf("HandleOPMLExportDialog: Generating OPML format (extension: %s)", ext)
		data, err = opml.Generate(localFeeds)
	}

	if err != nil {
		log.Printf("Error generating export data: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		response.JSON(w, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	// Write content to selected file
	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		log.Printf("Error writing file: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		response.JSON(w, map[string]interface{}{
			"error": "Failed to write file",
		})
		return
	}

	response.JSON(w, map[string]interface{}{
		"status":   "success",
		"filePath": filePath,
	})
}
