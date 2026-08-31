package opml

import (
	"context"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
	"MrRSS/internal/jsonimport"
	"MrRSS/internal/models"
	"MrRSS/internal/opml"
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
// @Description  Import RSS feed subscriptions from an OPML or JSON file. Accepts either a multipart upload or a raw request body.
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
		// Handle raw body upload. The filename hint lets the client pick JSON parsing.
		filename = r.URL.Query().Get("filename")
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

// HandleOPMLImportDialog is not available: the native client presents its own
// file picker and uploads through /api/opml/import.
// @Summary      Import dialog (not available)
// @Description  File dialogs are handled by the native client. Use /api/opml/import with a file upload instead
// @Tags         opml
// @Accept       json
// @Produce      json
// @Success      501  {object}  map[string]string  "Not implemented error"
// @Router       /opml/import/dialog [post]
func HandleOPMLImportDialog(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	log.Printf("File dialog operations are handled by the native client")
	w.WriteHeader(http.StatusNotImplemented)
	response.JSON(w, map[string]interface{}{
		"error": "File dialogs are handled by the native client. Use /api/opml/import with a file upload instead.",
	})
}

// HandleOPMLExportDialog is not available: the native client presents its own
// save panel and downloads through /api/opml/export.
// @Summary      Export dialog (not available)
// @Description  File dialogs are handled by the native client. Use /api/opml/export for a direct download instead
// @Tags         opml
// @Accept       json
// @Produce      json
// @Success      501  {object}  map[string]string  "Not implemented error"
// @Router       /opml/export/dialog [post]
func HandleOPMLExportDialog(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	log.Printf("File dialog is handled by the native client")
	w.WriteHeader(http.StatusNotImplemented)
	response.JSON(w, map[string]interface{}{
		"error": "File dialogs are handled by the native client. Use /api/opml/export for a direct download instead.",
	})
}
