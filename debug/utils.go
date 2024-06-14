package debug

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/DoomLordor/logger"
)

type LogLevel struct {
	ModuleName string `json:"module_name"`
	LogLevel   string `json:"log_level"`
}

func setLogLevel(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	logLevel := &LogLevel{}
	err := json.NewDecoder(r.Body).Decode(logLevel)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"status": "ok"}`)
		return
	}

	if logLevel.ModuleName == "" {
		logLevel.ModuleName = logger.BaseLoggerName
	}

	if logLevel.LogLevel == "" {
		logLevel.LogLevel = "info"
	}

	logger.SetLevel(logLevel.ModuleName, logLevel.LogLevel)

	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, `{"status": "ok"}`)
}

func healthCheckHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, `{"alive": true}`)
}
