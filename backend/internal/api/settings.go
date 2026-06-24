package api

import (
	"encoding/json"
	"net/http"

	"github.com/alpha-lens/backend/internal/models"
)

func (h *Handler) GetSettings(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.Settings)
}

func (h *Handler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var s models.Settings
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	*h.Settings = s
	writeJSON(w, http.StatusOK, h.Settings)
}
