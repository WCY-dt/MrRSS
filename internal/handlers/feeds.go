package handlers

import (
	"encoding/json"
	"net/http"
)

// HandleFeeds returns all feeds
func (h *Handler) HandleFeeds(w http.ResponseWriter, r *http.Request) {
	feeds, err := h.DB.GetFeeds()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, feeds)
}

// HandleAddFeed adds a new feed subscription
func (h *Handler) HandleAddFeed(w http.ResponseWriter, r *http.Request) {
	var req addFeedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.Fetcher.AddSubscription(req.URL, req.Category, req.Title); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondOK(w)
}

// HandleDeleteFeed deletes a feed by ID
func (h *Handler) HandleDeleteFeed(w http.ResponseWriter, r *http.Request) {
	id, err := parseInt64QueryParam(r, "id")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.DB.DeleteFeed(id); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondOK(w)
}

// HandleUpdateFeed updates feed information
func (h *Handler) HandleUpdateFeed(w http.ResponseWriter, r *http.Request) {
	var req updateFeedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.DB.UpdateFeed(req.ID, req.Title, req.URL, req.Category); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondOK(w)
}
