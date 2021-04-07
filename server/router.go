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
	router.Handler(http.MethodGet, "/", http.RedirectHandler("/index/", http.StatusTemporaryRedirect))
	router.GET("/index/", index(templates, ds))
	router.GET("/index/:id/edit", editBookmark(templates, ds))
	router.GET("/index/:id/view", viewBookmark(templates, ds))
	router.POST("/index/", submitNewBookmark(ds, index(templates, ds)))
	router.PUT("/index/:id", submitEditedBookmark(ds, viewBookmark(templates, ds)))
	router.DELETE("/index/:id", deleteBookmark(ds, index(templates, ds)))
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
