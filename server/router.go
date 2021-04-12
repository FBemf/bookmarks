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
const apiPrefix = "/api"

type sessionMiddleware = func(sessionHandler) httprouter.Handle
type sessionHandler = func(datastore.Session, http.ResponseWriter, *http.Request, httprouter.Params)

func MakeRouter(templates *templates.Templates, static fs.FS, ds *datastore.Datastore) http.Handler {
	router := httprouter.New()
	router.Handler(http.MethodGet, "/", http.RedirectHandler("/bookmarks/", http.StatusFound))
	router.GET("/login", loginPage(templates, ds))
	router.POST("/login", doLogin(templates, ds))
	router.GET("/logout", logout)

	router.OPTIONS(apiPrefix+"/newbookmark", corsOptions())
	router.POST(apiPrefix+"/newbookmark", apiNewBookmark(ds))
	router.GET(apiPrefix+"/export", apiExport(ds))

	router.ServeFiles("/static/*filepath", http.FS(static))

	routeProtected(router, templates, ds)
	return RequestLogger{router}
}

func routeProtected(router *httprouter.Router, templates *templates.Templates, ds *datastore.Datastore) {
	// Note: because we use same-site=lax cookies, all dangerous endpoints have to be POSTs
	// Also, all form endpoints (i.e. all POSTs) must have csrf middleware
	auth := auth(ds, "/login")
	router.GET(bookmarksPrefix, auth(index(templates, ds)))
	router.GET(bookmarksPrefix+"/edit/:id", auth(editBookmark(templates, ds)))
	router.GET(bookmarksPrefix+"/view/:id", auth(viewBookmark(templates, ds)))
	router.POST(bookmarksPrefix+"/create", auth(csrf(submitNewBookmark(ds))))
	router.POST(bookmarksPrefix+"/edit/:id", auth(csrf(submitEditedBookmark(ds))))
	router.POST(bookmarksPrefix+"/delete/:id", auth(csrf(deleteBookmark(ds))))

	router.GET(keysPrefix, auth(keys(templates, ds)))
	router.POST(keysPrefix+"/create", auth(csrf(createKey(templates, ds))))
	router.POST(keysPrefix+"/delete/:id", auth(csrf(deleteKey(templates, ds))))

	router.GET("/export", auth(export(templates, ds)))
	router.GET("/tags", auth(tags(templates, ds)))
}

type RequestLogger struct {
	h http.Handler
}

func (rl RequestLogger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)
	rl.h.ServeHTTP(w, r)
}

func auth(ds *datastore.Datastore, loginPath string) sessionMiddleware {
	redirecter := http.RedirectHandler(loginPath, http.StatusSeeOther)
	return func(h sessionHandler) httprouter.Handle {
		return func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
			session, valid, err := authenticateSession(ds, req)
			if err != nil {
				ErrorPage(resp, http.StatusInternalServerError)
			}
			if valid {
				h(session, resp, req, params)
			} else {
				redirecter.ServeHTTP(resp, req)
			}
		}
	}
}

func csrf(h sessionHandler) sessionHandler {
	return func(session datastore.Session, resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		req.ParseForm()
		csrf := req.Form.Get(templates.CsrfTokenName)
		if session.CsrfToken != csrf {
			ErrorPage(resp, http.StatusForbidden)
			return
		} else {
			h(session, resp, req, params)
		}
	}
}
