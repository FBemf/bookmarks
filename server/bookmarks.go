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
	Bookmarks    []datastore.Bookmark
	Pager        pager
	SearchParams urlparams.SearchParams
	NumBookmarks int64
	CsrfToken    string
}

type bookmarkData struct {
	Bookmark     datastore.Bookmark
	SearchParams urlparams.SearchParams
	CsrfToken    string
}

func index(templates *templates.Templates, ds *datastore.Datastore) sessionHandler {
	return func(session datastore.Session, resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		resp.Header().Set("Content-Type", "text/html; charset=UTF-8")
		urlParams, err := urlparams.GetQueryParams(req)
		if err != nil {
			ErrorPage(resp, http.StatusBadRequest)
			return
		}

		query := datastore.NewQueryInfo(pageSize)
		query.Offset = pageSize * uint(urlParams.Page-1)
		query.Search = urlParams.Search
		if urlParams.Order == urlparams.ReverseOrder {
			query.Reverse = true
		}
		query.Tags = urlParams.SearchTags

		bookmarks, err := ds.GetBookmarks(query)
		if err != nil {
			ErrorPage(resp, http.StatusInternalServerError)
			log.Printf("getting bookmarks: %v", err)
			return
		}

		numBookmarks, err := ds.GetNumBookmarks(query)
		if err != nil {
			ErrorPage(resp, http.StatusInternalServerError)
			log.Printf("getting number of bookmarks: %v", err)
			return
		}

		pager := createPager(urlParams.Page, int(numBookmarks+pageSize-1)/pageSize, pagerSideSize)
		err = templates.Index.ExecuteTemplate(resp, "base",
			indexData{bookmarks, pager, urlParams, numBookmarks, session.CsrfToken})
		if err != nil {
			ErrorPage(resp, http.StatusInternalServerError)
			log.Printf("writing template: %v", err)
			return
		}
	}
}

func viewBookmark(templates *templates.Templates, ds *datastore.Datastore) sessionHandler {
	return func(session datastore.Session, resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		resp.Header().Set("Content-Type", "text/html; charset=UTF-8")
		urlParams, err := urlparams.GetQueryParams(req)
		if err != nil {
			ErrorPage(resp, http.StatusBadRequest)
			return
		}

		bookmarkIdParam := params[0].Value
		bookmarkId, err := strconv.Atoi(bookmarkIdParam)
		if err != nil {
			ErrorPage(resp, http.StatusBadRequest)
			return
		}
		bookmark, err := ds.GetBookmark(int64(bookmarkId))
		if err != nil {
			ErrorPage(resp, http.StatusNotFound)
			return
		}
		err = templates.ViewBookmark.ExecuteTemplate(resp, "base", bookmarkData{bookmark, urlParams, session.CsrfToken})
		if err != nil {
			ErrorPage(resp, http.StatusInternalServerError)
			log.Printf("writing template: %v", err)
			return
		}
	}
}

func editBookmark(templates *templates.Templates, ds *datastore.Datastore) sessionHandler {
	return func(session datastore.Session, resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		resp.Header().Set("Content-Type", "text/html; charset=UTF-8")
		urlParams, err := urlparams.GetQueryParams(req)
		if err != nil {
			ErrorPage(resp, http.StatusBadRequest)
			return
		}

		bookmarkIdParam := params[0].Value
		bookmarkId, err := strconv.Atoi(bookmarkIdParam)
		if err != nil {
			ErrorPage(resp, http.StatusBadRequest)
			return
		}
		bookmark, err := ds.GetBookmark(int64(bookmarkId))
		if err != nil {
			ErrorPage(resp, http.StatusNotFound)
			return
		}
		err = templates.EditBookmark.ExecuteTemplate(resp, "base", bookmarkData{bookmark, urlParams, session.CsrfToken})
		if err != nil {
			ErrorPage(resp, http.StatusInternalServerError)
			log.Printf("writing template: %v", err)
			return
		}
	}
}

func submitEditedBookmark(ds *datastore.Datastore) sessionHandler {
	return func(session datastore.Session, resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		err := req.ParseForm()
		if err != nil {
			ErrorPage(resp, http.StatusBadRequest)
			return
		}
		name := req.Form.Get("name")
		url := req.Form.Get("url")
		description := req.Form.Get("description")
		if name == "" || url == "" {
			ErrorPage(resp, http.StatusBadRequest)
			return
		}
		tags := req.Form["tag"]

		bookmarkIdParam := params[0].Value
		id, err := strconv.Atoi(bookmarkIdParam)
		if err != nil {
			ErrorPage(resp, http.StatusBadRequest)
			return
		}
		if err != nil {
			ErrorPage(resp, http.StatusNotFound)
			return
		}

		err = req.ParseForm()
		if err != nil {
			ErrorPage(resp, http.StatusBadRequest)
			return
		}

		url = ensureProtocol(url)
		err = ds.UpdateBookmark(int64(id), name, url, description, tags)
		if err != nil {
			ErrorPage(resp, http.StatusInternalServerError)
			log.Printf("updating bookmark %d: %s", id, err)
			return
		}
		http.Redirect(resp, req, bookmarksPrefix+"/view/"+bookmarkIdParam, http.StatusSeeOther)
	}
}

func submitNewBookmark(ds *datastore.Datastore) sessionHandler {
	return func(session datastore.Session, resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		err := req.ParseForm()
		if err != nil {
			ErrorPage(resp, http.StatusBadRequest)
			return
		}
		name := req.Form.Get("name")
		url := req.Form.Get("url")
		description := req.Form.Get("description")
		if name == "" || url == "" {
			ErrorPage(resp, http.StatusBadRequest)
			return
		}
		url = ensureProtocol(url)
		tags := req.Form["tag"]

		err = ds.CreateBookmark(name, url, description, tags)
		if err != nil {
			ErrorPage(resp, http.StatusInternalServerError)
			log.Printf("adding new bookmark: %v", err)
			return
		}
		http.Redirect(resp, req, bookmarksPrefix, http.StatusSeeOther)
	}
}

func deleteBookmark(ds *datastore.Datastore) sessionHandler {
	return func(session datastore.Session, resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		bookmarkIdParam := params[0].Value
		id, err := strconv.Atoi(bookmarkIdParam)
		if err != nil {
			ErrorPage(resp, http.StatusBadRequest)
			return
		}
		err = ds.DeleteBookmark(int64(id))
		if err != nil {
			ErrorPage(resp, http.StatusInternalServerError)
			log.Printf("deleting bookmark %d: %v", id, err)
			return
		}
		http.Redirect(resp, req, bookmarksPrefix, http.StatusSeeOther)
	}
}

func ErrorPage(resp http.ResponseWriter, code int) {
	http.Error(resp, fmt.Sprintf("%d %s", code, http.StatusText(code)), code)
}

func ensureProtocol(url string) string {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}
	return url
}
