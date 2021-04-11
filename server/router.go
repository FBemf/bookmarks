package server

import (
	"io/fs"
	"local/bookmarks/datastore"
	"local/bookmarks/templates"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

const bookmarksPrefix = "/bookmarks"
const keysPrefix = "/keys"
const exportPrefix = "/export"
const apiPrefix = "/api"

type middleware = func(httprouter.Handle) httprouter.Handle

func MakeRouter(templates *templates.Templates, static fs.FS, ds *datastore.Datastore) http.Handler {
	router := httprouter.New()
	router.Handler(http.MethodGet, "/", http.RedirectHandler("/bookmarks/", http.StatusFound))
	router.GET("/login", loginPage(templates, ds))
	router.POST("/login", doLogin(templates, ds))
	router.GET("/logout", logout)

	router.POST(apiPrefix+"/newbookmark", apiNewBookmark(ds))
	router.GET(apiPrefix+"/export", apiExport(ds))

	router.ServeFiles("/static/*filepath", http.FS(static))

	routeProtected(router, templates, ds, auth(ds, "/login"))
	return RequestLogger{router}
}

func routeProtected(router *httprouter.Router, templates *templates.Templates, ds *datastore.Datastore, middleware middleware) {
	// note: because we use same-site=lax cookies for csrf protection,
	// all dangerous endpoints have to be POSTs
	router.GET(bookmarksPrefix, middleware(index(templates, ds)))
	router.GET(bookmarksPrefix+"/edit/:id", middleware(editBookmark(templates, ds)))
	router.GET(bookmarksPrefix+"/view/:id", middleware(viewBookmark(templates, ds)))
	router.POST(bookmarksPrefix+"/create", middleware(submitNewBookmark(ds)))
	router.POST(bookmarksPrefix+"/edit/:id", middleware(submitEditedBookmark(ds)))
	router.POST(bookmarksPrefix+"/delete/:id", middleware(deleteBookmark(ds)))

	router.GET(keysPrefix, middleware(keys(templates, ds)))
	router.POST(keysPrefix+"/create", middleware(createKey(templates, ds)))
	router.POST(keysPrefix+"/delete/:id", middleware(deleteKey(templates, ds)))

	router.GET(exportPrefix, middleware(export(templates, ds)))
}

type RequestLogger struct {
	h http.Handler
}

func (rl RequestLogger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)
	rl.h.ServeHTTP(w, r)
}

func auth(ds *datastore.Datastore, loginPath string) middleware {
	redirecter := http.RedirectHandler(loginPath, http.StatusSeeOther)
	return func(h httprouter.Handle) httprouter.Handle {
		return func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
			_, valid, err := authenticateSession(ds, req)
			if err != nil {
				ErrorPage(resp, http.StatusInternalServerError)
			}
			if valid {
				h(resp, req, params)
			} else {
				redirecter.ServeHTTP(resp, req)
			}
		}
	}
}
