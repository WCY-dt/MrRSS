package rules

import (
	"encoding/json"
	"net/http"

	"MrRSS/internal/handlers/core"
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
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var rule rules.Rule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(rule.Actions) == 0 {
		http.Error(w, "No actions specified", http.StatusBadRequest)
		return
	}

	engine := rules.NewEngine(h.DB)
	affected, err := engine.ApplyRule(rule)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		Success  bool `json:"success"`
		Affected int  `json:"affected"`
	}{
		Success:  true,
		Affected: affected,
	}
	json.NewEncoder(w).Encode(response)
}
