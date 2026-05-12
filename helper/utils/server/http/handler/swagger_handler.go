package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type SwaggerHandler interface {
	SwaggerIndex() http.HandlerFunc
	SwaggerDocs() http.HandlerFunc
}

type swaggerHandler struct {
	specs map[string]string
}

func NewSwaggerHandler() *swaggerHandler {
	h := &swaggerHandler{
		specs: make(map[string]string),
	}
	h.loadSpecs()
	return h
}

func (h *swaggerHandler) loadSpecs() {
	_ = filepath.Walk("api", func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(info.Name(), ".swagger.json") {
			parts := strings.Split(path, string(filepath.Separator))
			if len(parts) >= 2 {
				name := parts[1]
				h.specs[name] = path
			}
		}
		return nil
	})
}

func (h *swaggerHandler) SwaggerIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		html := `
<!DOCTYPE html>
<html>
<head>
  <title>Swagger UI</title>
  <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist/swagger-ui.css" />
  <link rel="icon" type="image/png" href="https://unpkg.com/swagger-ui-dist/favicon-32x32.png" sizes="32x32" />
  <link rel="icon" type="image/png" href="https://unpkg.com/swagger-ui-dist/favicon-16x16.png" sizes="16x16" />
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist/swagger-ui-bundle.js"></script>
<script src="https://unpkg.com/swagger-ui-dist/swagger-ui-standalone-preset.js"></script>
<script>
window.onload = function() {
  const ui = SwaggerUIBundle({
    dom_id: '#swagger-ui',
    urls: [
`
		// build list specs
		first := true
		for name := range h.specs {
			if !first {
				html += ","
			}
			html += `{url: "/swagger/docs/` + name + `", name: "` + name + `"}` + "\n"
			first = false
		}

		html += `
    ],
    "urls.primaryName": "` + firstKey(h.specs) + `",
    presets: [
      SwaggerUIBundle.presets.apis,
      SwaggerUIStandalonePreset
    ],
    layout: "StandaloneLayout"
  })
}
</script>
</body>
</html>
`
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(html))
	}
}

func (h *swaggerHandler) SwaggerDocs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/swagger/docs/")
		if name == "" {
			http.NotFound(w, r)
			return
		}
		if path, ok := h.specs[name]; ok {
			data, err := os.ReadFile(path)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(data)
			return
		}
		http.NotFound(w, r)
	}
}

func firstKey(m map[string]string) string {
	for k := range m {
		return k
	}
	return ""
}
