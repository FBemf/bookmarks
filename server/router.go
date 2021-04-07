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
	router.Handler(http.MethodGet, "/", http.RedirectHandler("/index/", http.StatusTemporaryRedirect))
	router.GET("/index", index(templates, ds))
	router.GET("/index/edit/:id", editBookmark(templates, ds))
	router.GET("/index/view/:id", viewBookmark(templates, ds))
	router.POST("/index/create", submitNewBookmark(ds, index(templates, ds)))
	router.POST("/index/edit/:id", submitEditedBookmark(ds, viewBookmark(templates, ds)))
	router.POST("/index/delete/:id", deleteBookmark(ds, index(templates, ds)))
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
