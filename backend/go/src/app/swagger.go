package app

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"strings"
)

//go:embed swagger/openapi.json
var openAPISpec []byte

const swaggerHTML = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1"/>
    <title>Game Master Notes API</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script>
      window.onload = () => {
        window.ui = SwaggerUIBundle({
          url: '/openapi.json',
          dom_id: '#swagger-ui',
          deepLinking: true,
          presets: [SwaggerUIBundle.presets.apis],
        });
      };
    </script>
  </body>
</html>`

func registerSwaggerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /openapi.json", func(w http.ResponseWriter, r *http.Request) {
		payload, err := withDynamicServerMetadata(openAPISpec, r)
		if err != nil {
			http.Error(w, "failed to render openapi document", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(payload)
	})

	mux.HandleFunc("GET /swagger", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(swaggerHTML))
	})

	mux.HandleFunc("GET /swagger/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger", http.StatusTemporaryRedirect)
	})
}

func withDynamicServerMetadata(spec []byte, r *http.Request) ([]byte, error) {
	var doc map[string]any
	if err := json.Unmarshal(spec, &doc); err != nil {
		return nil, err
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); forwarded != "" {
		scheme = forwarded
	}

	host := strings.TrimSpace(r.Host)
	if host == "" {
		host = "localhost"
	}

	doc["servers"] = []map[string]string{
		{
			"url":         scheme + "://" + host,
			"description": "Dynamic server inferred from the incoming request. Port can be configured via PORT env var.",
		},
	}

	return json.Marshal(doc)
}
