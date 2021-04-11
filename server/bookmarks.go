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
}

type bookmarkData struct {
	Bookmark     datastore.Bookmark
	SearchParams urlparams.SearchParams
}

func index(templates *templates.Templates, ds *datastore.Datastore) httprouter.Handle {
	return func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		urlParams, err := urlparams.GetUrlParams(req)
		if err != nil {
			ErrorPage(resp, http.StatusBadRequest)
			return
		}

		query := datastore.NewQueryInfo(PAGE_SIZE)
		query.Offset = PAGE_SIZE * uint(urlParams.Page-1)
		query.Search = urlParams.Search
		if urlParams.Order == urlparams.REVERSE_ORDER {
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

		pager := createPager(urlParams.Page, int(numBookmarks+PAGE_SIZE-1)/PAGE_SIZE, PAGER_SIDE_SIZE)
		err = templates.Index.ExecuteTemplate(resp, "base",
			indexData{bookmarks, pager, urlParams})
		if err != nil {
			ErrorPage(resp, http.StatusInternalServerError)
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
		err = templates.ViewBookmark.ExecuteTemplate(resp, "base", bookmarkData{bookmark, urlParams})
		if err != nil {
			ErrorPage(resp, http.StatusInternalServerError)
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
		err = templates.EditBookmark.ExecuteTemplate(resp, "base", bookmarkData{bookmark, urlParams})
		if err != nil {
			ErrorPage(resp, http.StatusInternalServerError)
			log.Printf("writing template: %v", err)
			return
		}

		resp.Header().Set("Content-Type", "text/html; charset=UTF-8")
	}
}

func submitEditedBookmark(ds *datastore.Datastore) httprouter.Handle {
	return func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
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
		name := req.Form.Get("name")
		url := req.Form.Get("url")
		description := req.Form.Get("description")
		if name == "" || url == "" {
			ErrorPage(resp, http.StatusBadRequest)
			return
		}
		tags := req.Form["tag"]

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

func submitNewBookmark(ds *datastore.Datastore) httprouter.Handle {
	return func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
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

func deleteBookmark(ds *datastore.Datastore) httprouter.Handle {
	return func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
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