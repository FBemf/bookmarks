package server

import (
	"fmt"
	"html/template"
	"local/bookmarks/datastore"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type indexData struct {
	Recent []datastore.Bookmark
}

func index(templates *template.Template, ds *datastore.Datastore) httprouter.Handle {
	return func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		resp.Header().Set("Content-Type", "text/html; charset=UTF-8")
		bookmarks, err := ds.RecentBookmarks(5)
		if err != nil {
			errorPage(resp, http.StatusInternalServerError)
			log.Printf("getting recent bookmarks: %v", err)
			return
		}
		err = templates.ExecuteTemplate(resp, "index.html",
			indexData{bookmarks})
		if err != nil {
			errorPage(resp, http.StatusInternalServerError)
			log.Printf("writing template: %v", err)
			return
		}
	}
}

func newBookmark(templates *template.Template, ds *datastore.Datastore, forward httprouter.Handle) httprouter.Handle {
	return func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		err := req.ParseForm()
		if err != nil {
			errorPage(resp, http.StatusBadRequest)
			return
		}
		name := req.Form.Get("name")
		url := req.Form.Get("url")
		description := req.Form.Get("description")
		if name == "" || url == "" {
			errorPage(resp, http.StatusBadRequest)
			return
		}
		ds.NewBookmark(name, url, description)
		forward(resp, req, params)
	}
}

func errorPage(resp http.ResponseWriter, code int) {
	http.Error(resp, fmt.Sprintf("%d %s", code, http.StatusText(code)), code)
}
