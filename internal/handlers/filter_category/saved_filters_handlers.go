package filter_category

import (
	"encoding/json"
	"net/http"
	"strconv"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/models"
)

// HandleSavedFilters handles CRUD operations for saved filters
// @Summary      Get all saved filters
// @Description  Retrieve all saved article filters
// @Tags         filters
// @Accept       json
// @Produce      json
// @Success      200  {array}   models.SavedFilter  "List of saved filters"
// @Router       /saved-filters [get]
// @Summary      Create a new saved filter
// @Description  Create a new article filter with custom conditions
// @Tags         filters
// @Accept       json
// @Produce      json
// @Param        request  body      object  true  "Filter details"
// @Success      201  {object}  models.SavedFilter  "Created filter"
// @Failure      400  {object}  map[string]string  "Bad request"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /saved-filters [post]
func HandleSavedFilters(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// GET: List all saved filters
		filters, err := h.DB.GetSavedFilters()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(filters)
		return
	}

	if r.Method == http.MethodPost {
		// POST: Create new saved filter
		var req struct {
			Name       string `json:"name"`
			Conditions string `json:"conditions"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Validate input
		if req.Name == "" {
			http.Error(w, "Filter name is required", http.StatusBadRequest)
			return
		}
		if req.Conditions == "" {
			http.Error(w, "Filter conditions are required", http.StatusBadRequest)
			return
		}

		// Check if filter with same name already exists
		existingFilters, err := h.DB.GetSavedFilters()
		if err != nil {
			http.Error(w, "Failed to check existing filters", http.StatusInternalServerError)
			return
		}

		for _, existing := range existingFilters {
			if existing.Name == req.Name {
				http.Error(w, `{"error": "A filter with this name already exists. Please choose a different name or edit the existing filter."}`, http.StatusConflict)
				return
			}
		}

		filter := &models.SavedFilter{
			Name:       req.Name,
			Conditions: req.Conditions,
			Position:   0, // Will be auto-assigned
		}

		id, err := h.DB.AddSavedFilter(filter)
		if err != nil {
			http.Error(w, "Failed to save filter: "+err.Error(), http.StatusInternalServerError)
			return
		}

		filter.ID = id
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(filter)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// HandleSavedFilter handles operations on a specific saved filter
// @Summary      Update a saved filter
// @Description  Update an existing saved filter's name or conditions
// @Tags         filters
// @Accept       json
// @Produce      json
// @Param        id  query      int  true  "Filter ID"
// @Param        request  body      object  true  "Updated filter details"
// @Success      200  {object}  map[string]string  "Success message"
// @Failure      400  {object}  map[string]string  "Bad request"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /saved-filters/filter [put]
// @Summary      Delete a saved filter
// @Description  Delete a saved filter by ID
// @Tags         filters
// @Param        id  query      int  true  "Filter ID"
// @Success      200  {object}  map[string]string  "Success message"
// @Failure      400  {object}  map[string]string  "Bad request"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /saved-filters/filter [delete]
func HandleSavedFilter(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid filter ID", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodPut || r.Method == http.MethodPatch {
		// UPDATE: Edit existing filter
		var req struct {
			Name       string `json:"name"`
			Conditions string `json:"conditions"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		filter := &models.SavedFilter{
			ID:         id,
			Name:       req.Name,
			Conditions: req.Conditions,
		}

		if err := h.DB.UpdateSavedFilter(filter); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		return
	}

	if r.Method == http.MethodDelete {
		// DELETE: Remove filter
		if err := h.DB.DeleteSavedFilter(id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// HandleReorderSavedFilters handles bulk reordering of saved filters
// @Summary      Reorder saved filters
// @Description  Update the position of multiple saved filters
// @Tags         filters
// @Accept       json
// @Produce      json
// @Param        request  body      []models.SavedFilter  true  "Filters with updated positions"
// @Success      200  {object}  map[string]string  "Success message"
// @Failure      400  {object}  map[string]string  "Bad request"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /saved-filters/reorder [post]
func HandleReorderSavedFilters(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var filters []models.SavedFilter
	if err := json.NewDecoder(r.Body).Decode(&filters); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.DB.ReorderSavedFilters(filters); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
