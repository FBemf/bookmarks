package server

import (
	"html/template"
	"io/fs"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func MakeRouter(templates *template.Template, static fs.FS) http.Handler {
	router := httprouter.New()
	router.Handler(http.MethodGet, "/", http.RedirectHandler("/index.html", http.StatusTemporaryRedirect))
	router.GET("/index.html", helloWorld(templates))
	router.ServeFiles("/static/*filepath", http.FS(static))
	return RequestLogger{router}
}

type RequestLogger struct {
	h http.Handler
}

func (rl RequestLogger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request %s %s", r.Method, r.URL.Path)
	rl.h.ServeHTTP(w, r)
}
