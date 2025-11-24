package handlers

import (
	"context"
	"io"
	"log"
	"net/http"
	"strings"

	"MrRSS/internal/opml"
)

// HandleOPMLImport imports feeds from an OPML file
func (h *Handler) HandleOPMLImport(w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleOPMLImport: ContentLength: %d", r.ContentLength)
	contentType := r.Header.Get("Content-Type")
	log.Printf("HandleOPMLImport: Content-Type: %s", contentType)

	var file io.Reader

	if strings.Contains(contentType, "multipart/form-data") {
		f, header, err := r.FormFile("file")
		if err != nil {
			log.Printf("Error getting form file: %v", err)
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		defer f.Close()
		log.Printf("HandleOPMLImport: Received file %s, size: %d", header.Filename, header.Size)

		if header.Size == 0 {
			respondWithError(w, http.StatusBadRequest, "Uploaded file is empty")
			return
		}
		file = f
	} else {
		// Handle raw body upload
		file = r.Body
		defer r.Body.Close()
	}

	feeds, err := opml.Parse(file)
	if err != nil {
		log.Printf("Error parsing OPML: %v", err)
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	go func() {
		for _, f := range feeds {
			h.Fetcher.ImportSubscription(f.Title, f.URL, f.Category)
		}
		h.Fetcher.FetchAll(context.Background())
	}()

	respondOK(w)
}

// HandleOPMLExport exports all feeds to an OPML file
func (h *Handler) HandleOPMLExport(w http.ResponseWriter, r *http.Request) {
	feeds, err := h.DB.GetFeeds()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	data, err := opml.Generate(feeds)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=subscriptions.opml")
	w.Header().Set("Content-Type", "text/xml")
	w.Write(data)
}
