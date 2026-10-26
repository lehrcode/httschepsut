package main

import (
	"fmt"
	"log/slog"
	"net/http"
)

type Param struct {
	Name string
	Kind string
}

type Level struct {
	ID          string
	Summary     string
	Description string
	PathSuffix  string
	Params      []Param
	Handler     func(w http.ResponseWriter, r *http.Request) int
}

func (l Level) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	slog.Debug(fmt.Sprintf("%s %s", r.Method, r.URL))
	var state = l.Handler(w, r)
	if stateStore.Get(l.ID) != StateSolved {
		stateStore.Put(l.ID, state)
		slog.Debug(fmt.Sprintf("  Level %s -> %s", l.ID, state))
	}
}
