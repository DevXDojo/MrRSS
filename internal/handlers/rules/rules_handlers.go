package rules

import (
	"encoding/json"
	"net/http"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
	"MrRSS/internal/rules"
)

// HandleApplyRule applies a rule to matching articles
// @Summary      Apply rule to articles
// @Description  Apply a rule with conditions and actions to matching articles (mark as read, favorite, etc.)
// @Tags         rules
// @Accept       json
// @Produce      json
// @Param        rule  body      rules.Rule  true  "Rule definition (conditions and actions)"
// @Success      200  {object}  map[string]interface{}  "Application result (success, affected count)"
// @Failure      400  {object}  map[string]string  "Bad request (invalid rule or no actions)"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /rules/apply [post]
func HandleApplyRule(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	var rule rules.Rule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		response.Error(w, err, http.StatusBadRequest)
		return
	}

	if len(rule.Actions) == 0 {
		response.Error(w, nil, http.StatusBadRequest)
		return
	}

	engine := rules.NewEngine(h.DB)
	affected, err := engine.ApplyRule(rule)
	if err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	response.JSON(w, map[string]interface{}{
		"success":  true,
		"affected": affected,
	})
}
