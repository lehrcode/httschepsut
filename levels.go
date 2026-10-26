package main

import (
	"fmt"
	"net/http"
)

const (
	StateUntouched = iota
	StateSolved
	StateFailed
)

var levels = []Level{
	{
		ID:          "1",
		Summary:     "GET",
		Description: "GET benutzen",
		Handler:     NewMethodOnlyHandler(http.MethodGet),
	},
	{
		ID:          "2",
		Summary:     "GET mit Query-Parameter",
		Description: "Rufe den Pfad mit GET auf und gebe den Query-Parameter 'limit' mit dem Wert '5' mit.",
		Params: []Param{
			Param{
				Name: "limit",
				Kind: "Query",
			},
		},
		Handler: Level2Handler,
	},
	{
		ID:          "3",
		Summary:     "GET mit Path-Parameter",
		Description: "Rufe den Pfad mit GET auf und gebe die Id 47 als Path-Parameter mit.",
		Params: []Param{
			Param{
				Name: "id",
				Kind: "Path",
			},
		},
		PathSuffix: "/{id}",
		Handler:    Level3Handler,
	},
	{
		ID:          "4",
		Summary:     "HEAD",
		Description: "Rufe den Pfad mit HEAD auf.",
		Handler:     NewMethodOnlyHandler(http.MethodHead),
	},
	{
		ID:          "5",
		Summary:     "POST",
		Description: "Rufe den Pfad mit POST auf.",
		Handler:     NewMethodOnlyHandler(http.MethodPost),
	},
}

func NewMethodOnlyHandler(httpMethod string) func(w http.ResponseWriter, r *http.Request) int {
	return func(w http.ResponseWriter, r *http.Request) int {
		if r.Method != httpMethod {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return StateFailed
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Well done!")
		return StateSolved
	}
}

func Level2Handler(w http.ResponseWriter, r *http.Request) int {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return StateFailed
	}
	w.WriteHeader(http.StatusOK)
	if r.URL.Query().Get("limit") != "5" {
		fmt.Fprintln(w, "Wrong!")
		return StateFailed
	}
	fmt.Fprintln(w, "Well done!")
	return StateSolved
}

func Level3Handler(w http.ResponseWriter, r *http.Request) int {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return StateFailed
	}
	w.WriteHeader(http.StatusOK)
	if r.PathValue("id") != "47" {
		fmt.Fprintln(w, "Wrong!")
		return StateFailed
	}
	fmt.Fprintln(w, "Well done!")
	return StateSolved
}
