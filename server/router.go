package server

import (
	"io/fs"
	"local/bookmarks/datastore"
	"local/bookmarks/templates"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func MakeRouter(templates *templates.Templates, static fs.FS, ds *datastore.Datastore) http.Handler {
	router := httprouter.New()
	router.Handler(http.MethodGet, "/", http.RedirectHandler("/bookmarks/", http.StatusTemporaryRedirect))
	router.GET("/bookmarks", index(templates, ds))
	router.GET("/bookmarks/edit/:id", editBookmark(templates, ds))
	router.GET("/bookmarks/view/:id", viewBookmark(templates, ds))
	router.POST("/bookmarks/create", submitNewBookmark(ds, index(templates, ds)))
	router.POST("/bookmarks/edit/:id", submitEditedBookmark(ds, viewBookmark(templates, ds)))
	router.POST("/bookmarks/delete/:id", deleteBookmark(ds, index(templates, ds)))
	router.ServeFiles("/static/*filepath", http.FS(static))
	return RequestLogger{router}
}

type RequestLogger struct {
	h http.Handler
}

func (rl RequestLogger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)
	rl.h.ServeHTTP(w, r)
}
