package server

import (
	"io/fs"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"local/bookmarks/datastore"
	"local/bookmarks/templates"
)

func MakeRouter(templates *templates.Templates, static fs.FS, ds *datastore.Datastore) http.Handler {
	router := httprouter.New()
	router.Handler(http.MethodGet, "/", http.RedirectHandler("/index.html", http.StatusTemporaryRedirect))
	router.GET("/index.html", index(templates, ds))
	router.GET("/edit/:id", editBookmark(templates, ds))
	router.GET("/bookmark/:id", viewBookmark(templates, ds))
	router.POST("/bookmark/", submitNewBookmark(templates, ds))
	router.PUT("/bookmark/:id", submitEditedBookmark(templates, ds))
	router.DELETE("/bookmark/:id", deleteBookmark(templates, ds))
	router.ServeFiles("/static/*filepath", http.FS(static))
	return RequestLogger{router}
}

type RequestLogger struct {
	h http.Handler
}

func (rl RequestLogger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)
	rl.h.ServeHTTP(w, r)
}
