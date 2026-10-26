package main

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/lehrcode/httschepsut/store"
)

type indexHandler struct {
	tmpl       *template.Template
	stateStore store.Store[int]
}

type levelWithState struct {
	Level
	State int
}

func (h *indexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	slog.Debug(fmt.Sprintf("%s %s", r.Method, r.URL))
	var levelsWithState = make([]levelWithState, 0, len(levels))
	var openLevelID string
	for _, level := range levels {
		var state = stateStore.Get(level.ID)
		if openLevelID == "" && state == StateUntouched {
			openLevelID = level.ID
		}
		levelsWithState = append(levelsWithState, levelWithState{Level: level, State: state})
	}
	var data = map[string]any{
		"BaseURL":     baseURL,
		"Levels":      levelsWithState,
		"OpenLevelID": openLevelID,
	}
	if err := h.tmpl.Execute(w, data); err != nil {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
