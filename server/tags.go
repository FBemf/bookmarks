package server

import (
	"local/bookmarks/datastore"
	"local/bookmarks/templates"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func tags(templates *templates.Templates, ds *datastore.Datastore) httprouter.Handle {
	return func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		resp.Header().Set("Content-Type", "text/html; charset=UTF-8")

		tags, err := ds.GetTags()
		if err != nil {
			ErrorPage(resp, http.StatusInternalServerError)
			log.Printf("getting tags: %v", err)
			return
		}

		err = templates.Tags.ExecuteTemplate(resp, "base", tags)
		if err != nil {
			ErrorPage(resp, http.StatusInternalServerError)
			log.Printf("writing template: %v", err)
			return
		}
	}
}
