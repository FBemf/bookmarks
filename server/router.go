package server

import (
	"io/fs"
	"local/bookmarks/datastore"
	"local/bookmarks/templates"
	"log"
	"net/http"
	"net/url"

	"github.com/julienschmidt/httprouter"
)

const bookmarksPrefix = "/bookmarks"
const keysPrefix = "/keys"
const apiPrefix = "/api"
const loginPrefix = "/login"

type sessionMiddleware = func(sessionHandler) httprouter.Handle
type sessionHandler = func(datastore.Session, http.ResponseWriter, *http.Request, httprouter.Params)

func MakeRouter(templates *templates.Templates, static fs.FS, ds *datastore.Datastore) http.Handler {
	router := httprouter.New()
	router.Handler(http.MethodGet, "/", http.RedirectHandler("/bookmarks/", http.StatusFound))
	router.GET(loginPrefix, loginPage(templates, ds))
	router.POST(loginPrefix, doLogin(templates, ds))
	router.GET("/logout", logout)

	router.OPTIONS(apiPrefix+"/bookmark", corsOptions([]string{http.MethodPost, http.MethodOptions}))
	router.POST(apiPrefix+"/bookmark", apiNewBookmark(ds))

	router.GET(apiPrefix+"/export", apiExport(ds))

	router.ServeFiles("/static/*filepath", http.FS(static))

	routeProtected(router, templates, ds)

	return RequestLogger{
		SecureHeadersMiddleware{router},
	}
}

func routeProtected(router *httprouter.Router, templates *templates.Templates, ds *datastore.Datastore) {
	auth := auth(ds, loginPrefix)

	GET := func(path string, handler sessionHandler) {
		router.GET(path, auth(handler))
	}
	POST := func(path string, handler sessionHandler) {
		router.POST(path, auth(csrf(handler)))
	}

	// Note: because we use same-site=lax cookies, dangerous endpoints have to be POSTs.
	// This endpoint is an exception because it *also* requires an api key to be passed as a url parameter.
	// (I know, I know, I'd rather pass it as a header too, but the bookmarklet can't do that. It's https-only)
	GET("/_bookmarklet", addFromBookmarklet(ds))

	GET(bookmarksPrefix, index(templates, ds))
	GET(bookmarksPrefix+"/edit/:id", editBookmark(templates, ds))
	GET(bookmarksPrefix+"/view/:id", viewBookmark(templates, ds))
	POST(bookmarksPrefix+"/create", submitNewBookmark(ds))
	POST(bookmarksPrefix+"/edit/:id", submitEditedBookmark(ds))
	POST(bookmarksPrefix+"/delete/:id", deleteBookmark(ds))

	GET(keysPrefix, keys(templates, ds))
	POST(keysPrefix+"/create", createKey(templates, ds))
	POST(keysPrefix+"/delete/:id", deleteKey(templates, ds))

	GET("/export", export(templates, ds))
	GET("/tags", tags(templates, ds))
}

type RequestLogger struct {
	h http.Handler
}

func (rl RequestLogger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)
	rl.h.ServeHTTP(w, r)
}

type SecureHeadersMiddleware struct {
	h http.Handler
}

func (m SecureHeadersMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
	w.Header().Set("X-Frame-Options", "DENY")
	m.h.ServeHTTP(w, r)
}

func auth(ds *datastore.Datastore, loginPath string) sessionMiddleware {
	loginUrl, err := url.Parse(loginPath)
	if err != nil {
		log.Panicf("illegal login path %s passed to auth: %s", loginPath, err)
	}
	return func(h sessionHandler) httprouter.Handle {
		return func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
			session, valid, err := authenticateSession(ds, req)
			if err != nil {
				ErrorPage(resp, http.StatusInternalServerError)
			}
			if valid {
				h(session, resp, req, params)
			} else {
				escapedReturnPath := url.QueryEscape(req.URL.String())
				redirectUrl := loginUrl
				q := redirectUrl.Query()
				q.Set("redirectTo", escapedReturnPath)
				redirectUrl.RawQuery = q.Encode()
				http.Redirect(resp, req, redirectUrl.String(), http.StatusSeeOther)
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
