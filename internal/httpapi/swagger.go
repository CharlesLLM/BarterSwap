package httpapi

import (
	_ "embed"
	"log"
	"net/http"
)

var openAPISpec []byte

const swaggerPage = `<!doctype html>
<html lang="fr">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>BarterSwap API</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    SwaggerUIBundle({
      url: "/openapi.yaml",
      dom_id: "#swagger-ui",
      deepLinking: true,
      persistAuthorization: true
    });
  </script>
</body>
</html>`

func openAPIHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		methodNotAllowed(responseWriter, http.MethodGet)
		return
	}
	responseWriter.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	if _, err := responseWriter.Write(openAPISpec); err != nil {
		log.Printf("écriture du schéma OpenAPI : %v", err)
	}
}

func swaggerRedirectHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if request.URL.Path != "/swagger" {
		writeError(responseWriter, http.StatusNotFound, "route introuvable")
		return
	}
	if request.Method != http.MethodGet {
		methodNotAllowed(responseWriter, http.MethodGet)
		return
	}
	http.Redirect(responseWriter, request, "/swagger/", http.StatusMovedPermanently)
}

func swaggerHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		methodNotAllowed(responseWriter, http.MethodGet)
		return
	}
	responseWriter.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := responseWriter.Write([]byte(swaggerPage)); err != nil {
		log.Printf("écriture de Swagger UI : %v", err)
	}
}
