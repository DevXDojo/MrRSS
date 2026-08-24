package tags

import (
	"encoding/json"
	"net/http"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
	"MrRSS/internal/models"
)

// HandleTags handles GET and POST requests for tags.
// @Summary      List or create tags
// @Description  GET: Retrieve all tags. POST: Create a new tag with name and color.
// @Tags         tags
// @Accept       json
// @Produce      json
// @Param        request  body      object  false  "Tag details (for POST)"
// @Success      200  {array}   models.Tag  "List of tags (GET)"
// @Success      201  {object}  models.Tag  "Created tag (POST)"
// @Failure      400  {object}  map[string]string  "Bad request"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /tags [get]
// @Router       /tags [post]
func HandleTags(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// GET: Return all tags
		tags, err := h.DB.GetTags()
		if err != nil {
			response.Error(w, err, http.StatusInternalServerError)
			return
		}

		response.JSON(w, tags)
		return
	}

	if r.Method == http.MethodPost {
		// POST: Create new tag
		var req struct {
			Name  string `json:"name"`
			Color string `json:"color"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.Error(w, err, http.StatusBadRequest)
			return
		}

		// Validate input
		if req.Name == "" {
			response.Error(w, nil, http.StatusBadRequest)
			return
		}
		if req.Color == "" {
			req.Color = "#3B82F6" // Default blue color
		}

		tag := &models.Tag{
			Name:  req.Name,
			Color: req.Color,
		}

		id, err := h.DB.AddTag(tag)
		if err != nil {
			response.Error(w, err, http.StatusInternalServerError)
			return
		}

		tag.ID = id
		w.WriteHeader(http.StatusCreated)
		response.JSON(w, tag)
		return
	}

	response.Error(w, nil, http.StatusMethodNotAllowed)
}

// HandleTagUpdate updates an existing tag.
// @Summary      Update a tag
// @Description  Update an existing tag's name, color, and position
// @Tags         tags
// @Accept       json
// @Produce      json
// @Param        request  body      object  true  "Tag update details (id, name, color, position)"
// @Success      200  {object}  models.Tag  "Updated tag"
// @Failure      400  {object}  map[string]string  "Bad request"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /tags/update [post]
func HandleTagUpdate(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID       int64  `json:"id"`
		Name     string `json:"name"`
		Color    string `json:"color"`
		Position int    `json:"position"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, err, http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Name == "" {
		response.Error(w, nil, http.StatusBadRequest)
		return
	}
	if req.Color == "" {
		response.Error(w, nil, http.StatusBadRequest)
		return
	}

	err := h.DB.UpdateTag(req.ID, req.Name, req.Color, req.Position)
	if err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	// Return the updated tag
	tag := &models.Tag{
		ID:       req.ID,
		Name:     req.Name,
		Color:    req.Color,
		Position: req.Position,
	}
	response.JSON(w, tag)
}

// HandleTagDelete deletes a tag by ID.
// @Summary      Delete a tag
// @Description  Delete an existing tag by its ID
// @Tags         tags
// @Accept       json
// @Produce      json
// @Param        request  body      object  true  "Tag ID to delete"
// @Success      200  {object}  map[string]string  "Deletion status"
// @Failure      400  {object}  map[string]string  "Bad request"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /tags/delete [post]
func HandleTagDelete(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, err, http.StatusBadRequest)
		return
	}

	err := h.DB.DeleteTag(req.ID)
	if err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	response.JSON(w, map[string]string{"status": "deleted"})
}

// HandleTagReorder changes the position of a tag.
// @Summary      Reorder a tag
// @Description  Change the display position of a tag in the list
// @Tags         tags
// @Accept       json
// @Produce      json
// @Param        request  body      object  true  "Tag ID and new position"
// @Success      200  {array}   models.Tag  "Updated list of all tags"
// @Failure      400  {object}  map[string]string  "Bad request"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /tags/reorder [post]
func HandleTagReorder(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID          int64 `json:"id"`
		NewPosition int   `json:"new_position"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, err, http.StatusBadRequest)
		return
	}

	err := h.DB.ReorderTag(req.ID, req.NewPosition)
	if err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	// Return updated list of tags
	tags, err := h.DB.GetTags()
	if err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	response.JSON(w, tags)
}

// RegisterTagRoutes registers all tag-related routes.
func RegisterTagRoutes(h *core.Handler, path string) {
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		HandleTags(h, w, r)
	})
	http.HandleFunc(path+"/update", func(w http.ResponseWriter, r *http.Request) {
		HandleTagUpdate(h, w, r)
	})
	http.HandleFunc(path+"/delete", func(w http.ResponseWriter, r *http.Request) {
		HandleTagDelete(h, w, r)
	})
	http.HandleFunc(path+"/reorder", func(w http.ResponseWriter, r *http.Request) {
		HandleTagReorder(h, w, r)
	})
}
