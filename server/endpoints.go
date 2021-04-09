package server

import (
	"fmt"
	"local/bookmarks/datastore"
	"local/bookmarks/templates"
	"local/bookmarks/urlparams"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
)

type indexData struct {
	Bookmarks []datastore.Bookmark
	Pager     pager
	UrlParams urlparams.UrlParams
}

type bookmarkData struct {
	Bookmark  datastore.Bookmark
	UrlParams urlparams.UrlParams
}

func index(templates *templates.Templates, ds *datastore.Datastore) httprouter.Handle {
	return func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		urlParams, err := urlparams.GetUrlParams(req)
		if err != nil {
			errorPage(resp, http.StatusBadRequest)
			return
		}

		bookmarks, err := ds.GetRecentBookmarks(PAGE_SIZE, PAGE_SIZE*uint(urlParams.Page-1))
		if err != nil {
			errorPage(resp, http.StatusInternalServerError)
			log.Printf("getting recent bookmarks: %v", err)
			return
		}
		numPages, err := ds.GetNumBookmarks()
		if err != nil {
			errorPage(resp, http.StatusInternalServerError)
			log.Printf("getting number of recent bookmarks: %v", err)
			return
		}

		pager := createPager(urlParams.Page, int(numPages+PAGE_SIZE-1)/PAGE_SIZE, PAGER_SIDE_SIZE)
		err = templates.Index.ExecuteTemplate(resp, "base",
			indexData{bookmarks, pager, urlParams})
		if err != nil {
			errorPage(resp, http.StatusInternalServerError)
			log.Printf("writing template: %v", err)
			return
		}

		resp.Header().Set("Content-Type", "text/html; charset=UTF-8")
	}
}

func viewBookmark(templates *templates.Templates, ds *datastore.Datastore) httprouter.Handle {
	return func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		urlParams, err := urlparams.GetUrlParams(req)
		if err != nil {
			errorPage(resp, http.StatusBadRequest)
			return
		}

		bookmarkIdParam := params[0].Value
		bookmarkId, err := strconv.Atoi(bookmarkIdParam)
		if err != nil {
			errorPage(resp, http.StatusBadRequest)
			return
		}
		bookmark, err := ds.GetBookmark(bookmarkId)
		if err != nil {
			errorPage(resp, http.StatusNotFound)
			return
		}
		err = templates.ViewBookmark.ExecuteTemplate(resp, "base", bookmarkData{bookmark, urlParams})
		if err != nil {
			errorPage(resp, http.StatusInternalServerError)
			log.Printf("writing template: %v", err)
			return
		}

		resp.Header().Set("Content-Type", "text/html; charset=UTF-8")
	}
}

func editBookmark(templates *templates.Templates, ds *datastore.Datastore) httprouter.Handle {
	return func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		urlParams, err := urlparams.GetUrlParams(req)
		if err != nil {
			errorPage(resp, http.StatusBadRequest)
			return
		}

		bookmarkIdParam := params[0].Value
		bookmarkId, err := strconv.Atoi(bookmarkIdParam)
		if err != nil {
			errorPage(resp, http.StatusBadRequest)
			return
		}
		bookmark, err := ds.GetBookmark(bookmarkId)
		if err != nil {
			errorPage(resp, http.StatusNotFound)
			return
		}
		err = templates.EditBookmark.ExecuteTemplate(resp, "base", bookmarkData{bookmark, urlParams})
		if err != nil {
			errorPage(resp, http.StatusInternalServerError)
			log.Printf("writing template: %v", err)
			return
		}

		resp.Header().Set("Content-Type", "text/html; charset=UTF-8")
	}
}

func submitEditedBookmark(ds *datastore.Datastore, forward httprouter.Handle) httprouter.Handle {
	return func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		bookmarkIdParam := params[0].Value
		id, err := strconv.Atoi(bookmarkIdParam)
		if err != nil {
			errorPage(resp, http.StatusBadRequest)
			return
		}
		if err != nil {
			errorPage(resp, http.StatusNotFound)
			return
		}
		err = req.ParseForm()
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
		url = ensureProtocol(url)
		err = ds.UpdateBookmark(id, name, url, description)
		if err != nil {
			errorPage(resp, http.StatusInternalServerError)
			log.Printf("updating bookmark %d: %s", id, err)
			return
		}
		forward(resp, req, params)
	}
}

func submitNewBookmark(ds *datastore.Datastore, forward httprouter.Handle) httprouter.Handle {
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
		url = ensureProtocol(url)
		err = ds.CreateBookmark(name, url, description)
		if err != nil {
			errorPage(resp, http.StatusInternalServerError)
			log.Printf("adding new bookmark: %v", err)
			return
		}
		forward(resp, req, params)
	}
}

func deleteBookmark(ds *datastore.Datastore, forward httprouter.Handle) httprouter.Handle {
	return func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		bookmarkIdParam := params[0].Value
		id, err := strconv.Atoi(bookmarkIdParam)
		if err != nil {
			errorPage(resp, http.StatusBadRequest)
			return
		}
		err = ds.DeleteBookmark(id)
		if err != nil {
			errorPage(resp, http.StatusInternalServerError)
			log.Printf("deleting template %d: %v", id, err)
			return
		}
		forward(resp, req, params)
	}
}

func errorPage(resp http.ResponseWriter, code int) {
	http.Error(resp, fmt.Sprintf("%d %s", code, http.StatusText(code)), code)
}

func ensureProtocol(url string) string {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}
	return url
}
