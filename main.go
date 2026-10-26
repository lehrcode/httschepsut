package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/lehrcode/httschepsut/store"
	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"
)

var (
	//go:embed index.gohtml
	indexTemplate string
	port          int
	debug         bool
	stateStore    = store.New[int]()
	baseURL       string
)

func main() {
	flag.IntVar(&port, "port", 8080, "HTTP server port")
	flag.BoolVar(&debug, "debug", false, "debug mode")
	flag.Parse()

	if debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	if port == 80 {
		baseURL = "http://localhost"
	} else {
		baseURL = fmt.Sprintf("http://localhost:%d", port)
	}

	var (
		mdp = parser.New()
		mdr = html.New()
	)

	http.Handle("/{$}", &indexHandler{
		tmpl: template.Must(template.New("index").
			Funcs(template.FuncMap{"md": func(src string) template.HTML {
				var buf bytes.Buffer
				node := mdp.ParseStringSource(src)
				if err := mdr.RenderStringSource(&buf, src, node); err != nil {
					slog.Error(err.Error())
				}
				return template.HTML(buf.String())
			}}).
			Parse(indexTemplate)),
	})

	http.HandleFunc("GET /events", sseHandler)

	for _, level := range levels {
		http.Handle("/level"+level.ID+level.PathSuffix, level)
		if level.PathSuffix != "" {
			http.Handle("/level"+level.ID, level)
			http.Handle("/level"+level.ID+"/", level)
		}
	}

	slog.Info(fmt.Sprintf("Starting server on http://localhost:%d/", port))
	if err := http.ListenAndServe(fmt.Sprintf("localhost:%d", port), nil); err != nil {
		slog.Error(err.Error())
	}
}
