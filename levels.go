package main

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"slices"
	"strings"
	"time"
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
		Description: "Rufe die URL mit **GET** auf.",
		HttpMethod:  http.MethodGet,
		Handler:     NewMethodOnlyHandler(http.MethodGet),
	},
	{
		ID:          "2",
		Summary:     "GET mit Query-Parameter",
		Description: "Rufe die URL mit **GET** auf und gebe den Query-Parameter `limit` mit dem Wert `5` mit.",
		HttpMethod:  http.MethodGet,
		Params: []Param{
			{
				Name:     "limit",
				Location: "Query",
				Type:     "Integer",
			},
		},
		Handler: Level2Handler,
	},
	{
		ID:          "3",
		Summary:     "GET mit Path-Parameter",
		Description: "Rufe das Dokument mit der Id `47` mit **GET** auf.",
		HttpMethod:  http.MethodGet,
		Params: []Param{
			{
				Name:     "id",
				Location: "Path",
				Type:     "Integer",
			},
		},
		PathSuffix: "/documents/{id}",
		Handler:    Level3Handler,
	},
	{
		ID:          "4",
		Summary:     "HEAD",
		Description: "Rufe die URL mit **HEAD** auf.",
		HttpMethod:  http.MethodHead,
		Handler:     NewMethodOnlyHandler(http.MethodHead),
	},
	{
		ID:          "5",
		Summary:     "HEAD mit Path-Parameter und Request-Header",
		Description: "Prüfe durch einen Aufruf mit **HEAD** ob sich das Dokument mit der Id `7` seit dem 01.10.2026 12:00 Uhr geändert hat.\n\nHinweis: Das Datum muss RFC1123-Format als '`Thu, 01 Oct 2026 12:00:00 GMT`' angegeben werden.",
		HttpMethod:  http.MethodHead,
		Params: []Param{
			{
				Name:     "id",
				Location: "Path",
			},
			{
				Name:     "If-Modified-Since",
				Location: "Header",
			},
		},
		PathSuffix: "/documents/{id}",
		Handler:    Level5Handler,
	},
	{
		ID:          "6",
		Summary:     "POST",
		Description: "Rufe die URL mit **POST** auf.",
		HttpMethod:  http.MethodPost,
		Handler:     NewMethodOnlyHandler(http.MethodPost),
	},
	{
		ID:          "7",
		Summary:     "POST mit Form-Parametern",
		Description: "Übermittle an die URL mit **POST** die Werte **name**=`Jim` und **email**=`jim@example.org` als Form-Parameter.",
		HttpMethod:  http.MethodPost,
		Params: []Param{
			{
				Name:     "Content-Type",
				Location: "Header",
			},
			{
				Name:     "name",
				Location: "Request-Body (Form)",
			},
			{
				Name:     "email",
				Location: "Request-Body (Form)",
			},
		},
		Handler: Level7Handler,
	},
	{
		ID:          "8",
		Summary:     "PUT mit JSON-Payload",
		Description: "Übermittle Daten im JSON-Format mit **PUT** an die URL.",
		HttpMethod:  http.MethodPut,
		Params: []Param{
			{
				Name:     "Content-Type",
				Location: "Header",
			},
		},
		Handler: Level8Handler,
	},
	{
		ID:          "9",
		Summary:     "PUT mit Basic-Auth",
		Description: "Überschreibe die Daten des Customers `11` mit **PUT** an die URL und verwende den Benutzernamen `admin` mit dem Passwort `trustno1`.\n\nHinweis: Benutzername und Passwort müssen mit einem Doppelpunkt verbunden und dann in Base64 umgewandelt werden (`YWRtaW46dHJ1c3RubzE=`).",
		HttpMethod:  http.MethodPut,
		Params: []Param{
			{
				Name:     "id",
				Location: "Path",
			},
			{
				Name:     "Authorization",
				Location: "Header",
			},
		},
		PathSuffix: "/customers/{id}",
		Handler:    Level9Handler,
	},
	{
		ID:          "10",
		Summary:     "DELETE mit Path-Parameter",
		Description: "Lösche mit **DELETE** den Customer mit der Id `23`.",
		HttpMethod:  http.MethodDelete,
		Params: []Param{
			{
				Name:     "id",
				Location: "Path",
				Type:     "Integer",
			},
		},
		PathSuffix: "/customers/{id}",
		Handler:    Level10Handler,
	},
	{
		ID:          "11",
		Summary:     "DELETE mit Query-Parametern",
		Description: "Lösche mit **DELETE** die Dokumente mit den Ids `23`, `31`, `42` mit einem Request.",
		HttpMethod:  http.MethodDelete,
		Params: []Param{
			{
				Name:     "id",
				Location: "Query",
				Type:     "Integer",
			},
		},
		PathSuffix: "/documents",
		Handler:    Level11Handler,
	},
	{
		ID:          "12",
		Summary:     "DELETE mit Path-Parameter und Basic-Auth",
		Description: "Lösche mit **DELETE** das Dokument mit der Id `14`. Verwende für die Authentifizierung den Benutzernamen `admin` und das Passwort `trustno1`.\n\nHinweis: Benutzername und Passwort müssen mit einem Doppelpunkt verbunden und dann in Base64 umgewandelt werden (`YWRtaW46dHJ1c3RubzE=`).",
		HttpMethod:  http.MethodDelete,
		Params: []Param{
			{
				Name:     "id",
				Location: "Path",
				Type:     "Integer",
			},
			{
				Name:     "Authorization",
				Location: "Header",
			},
		},
		PathSuffix: "/documents/{id}",
		Handler:    Level12Handler,
	},
	{
		ID:          "13",
		Summary:     "OPTIONS",
		Description: "Rufe die URL mit **OPTIONS** auf.",
		HttpMethod:  http.MethodOptions,
		Handler:     NewMethodOnlyHandler(http.MethodOptions),
	},
	{
		ID:          "14",
		Summary:     "PATCH mit JSON Merge Patch",
		Description: "Füge mit **PATCH** das Property **email** mit dem Wert `jim@example.org` zu dem Customer mit der Id `42` hinzu.",
		HttpMethod:  http.MethodPatch,
		Params: []Param{
			{
				Name:     "id",
				Location: "Path",
				Type:     "Integer",
			},
			{
				Name:     "Content-Type",
				Location: "Header",
			},
			{
				Name:     "email",
				Location: "Request-Body (JSON)",
			},
		},
		PathSuffix: "/customers/{id}",
		Handler:    Level14Handler,
	},
	{
		ID:          "15",
		Summary:     "QUERY mit Form-Parameter",
		Description: "Suche mit einem **QUERY**-Request an die URL nach dem Suchbegriff \"Hatschepsut\". Übergebe den Suchbegriff mit dem Property **q** in einem JSON-Dokument im Request-Body.",
		HttpMethod:  "QUERY",
		Params: []Param{
			{
				Name:     "Content-Type",
				Location: "Header",
			},
			{
				Name:     "q",
				Location: "Request-Body (JSON)",
			},
		},
		PathSuffix: "/search",
		Handler:    Level15Handler,
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
	if r.URL.Query().Get("limit") != "5" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return StateFailed
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Well done!")
	return StateSolved
}

func Level3Handler(w http.ResponseWriter, r *http.Request) int {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return StateFailed
	}
	if r.PathValue("id") != "47" {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return StateFailed
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Well done!")
	return StateSolved
}

func Level5Handler(w http.ResponseWriter, r *http.Request) int {
	if r.Method != http.MethodHead {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return StateFailed
	}
	if r.PathValue("id") != "7" {
		w.WriteHeader(http.StatusNotFound)
		return StateFailed
	}

	if t, err := time.Parse(time.RFC1123, r.Header.Get("If-Modified-Since")); err != nil || t.Format(time.DateTime) != "2026-10-01 12:00:00" {
		w.WriteHeader(http.StatusBadRequest)
		return StateFailed
	}

	w.Header().Set("Date", time.Now().UTC().Format(time.RFC1123))
	w.Header().Set("Last-Modified", time.Date(2026, time.January, 7, 13, 32, 4, 0, time.UTC).Format(time.RFC1123))
	w.WriteHeader(http.StatusNotModified)

	return StateSolved
}

func Level7Handler(w http.ResponseWriter, r *http.Request) int {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return StateFailed
	}
	if mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil ||
		mediaType != "application/x-www-form-urlencoded" {
		http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
		return StateFailed
	}

	if !strings.EqualFold(strings.TrimSpace(r.FormValue("name")), "Jim") ||
		!strings.EqualFold(strings.TrimSpace(r.FormValue("email")), "jim@example.org") {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return StateFailed
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Well done!")
	return StateSolved
}

func Level8Handler(w http.ResponseWriter, r *http.Request) int {
	if r.Method != http.MethodPut {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return StateFailed
	}
	if mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil ||
		mediaType != "application/json" {
		http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
		return StateFailed
	}

	var jsonPayload any

	jsonBytes, _ := io.ReadAll(r.Body)
	if err := json.Unmarshal(jsonBytes, &jsonPayload); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return StateFailed
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Well done!")
	return StateSolved
}

func Level9Handler(w http.ResponseWriter, r *http.Request) int {
	if r.Method != http.MethodPut {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return StateFailed
	}
	if usr, pwd, ok := r.BasicAuth(); ok {
		if !(usr == "admin" && pwd == "trustno1") {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return StateFailed
		}
	} else {
		w.Header().Set("WWW-Authenticate", `Basic`)
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return StateFailed
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Well done!")
	return StateSolved
}

func Level10Handler(w http.ResponseWriter, r *http.Request) int {
	if r.Method != http.MethodDelete {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return StateFailed
	}
	if r.PathValue("id") != "23" {
		http.Error(w, http.StatusText(http.StatusGone), http.StatusGone)
		return StateFailed
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Well done!")
	return StateSolved
}

func Level11Handler(w http.ResponseWriter, r *http.Request) int {
	if r.Method != http.MethodDelete {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return StateFailed
	}
	var params = r.URL.Query()["id"]
	slices.Sort(params)
	if !slices.Equal(params, []string{"23", "31", "42"}) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return StateFailed
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Well done!")
	return StateSolved
}

func Level12Handler(w http.ResponseWriter, r *http.Request) int {
	if r.Method != http.MethodDelete {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return StateFailed
	}
	if usr, pwd, ok := r.BasicAuth(); ok {
		if !(usr == "admin" && pwd == "trustno1") {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return StateFailed
		}
	} else {
		w.Header().Set("WWW-Authenticate", `Basic`)
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return StateFailed
	}
	if r.PathValue("id") != "14" {
		http.Error(w, http.StatusText(http.StatusGone), http.StatusGone)
		return StateFailed
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Well done!")
	return StateSolved
}

func Level14Handler(w http.ResponseWriter, r *http.Request) int {
	if r.Method != http.MethodPatch {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return StateFailed
	}
	if mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil ||
		(mediaType != "application/merge-patch+json" && mediaType != "application/json") {
		http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
		return StateFailed
	}
	if r.PathValue("id") != "42" {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return StateFailed
	}

	var jsonPayload map[string]string

	jsonBytes, _ := io.ReadAll(r.Body)
	if err := json.Unmarshal(jsonBytes, &jsonPayload); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return StateFailed
	}

	if !strings.EqualFold(strings.TrimSpace(jsonPayload["email"]), "jim@example.org") {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return StateFailed
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Well done!")
	return StateSolved
}

func Level15Handler(w http.ResponseWriter, r *http.Request) int {
	if r.Method != "QUERY" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return StateFailed
	}
	if mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil ||
		mediaType != "application/json" {
		http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
		return StateFailed
	}

	var jsonPayload map[string]string

	jsonBytes, _ := io.ReadAll(r.Body)
	if err := json.Unmarshal(jsonBytes, &jsonPayload); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return StateFailed
	}

	if !strings.EqualFold(strings.TrimSpace(jsonPayload["q"]), "Hatschepsut") {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return StateFailed
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Well done!")
	return StateSolved
}
