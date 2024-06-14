package rest

import (
	"io"
	"net/http"
)

func notFound(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	_, _ = io.WriteString(w, `{"error": "url not found"}`)
}
