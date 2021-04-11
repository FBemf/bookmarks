package server

import (
	"local/bookmarks/datastore"
	"local/bookmarks/templates"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func export(templates *templates.Templates, ds *datastore.Datastore) httprouter.Handle {
	return func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		resp.Header().Set("Content-Type", "text/html; charset=UTF-8")
		req.ParseForm()
		really := req.Form.Get("really")
		var exportData string
		var err error
		if really == "yes" {
			exportBytes, err := ds.Export()
			if err != nil {
				ErrorPage(resp, http.StatusInternalServerError)
			}
			exportData = string(exportBytes)
		} else {
			exportData = "Press \"Export Data\" to populate this field with your data."
		}

		err = templates.Export.ExecuteTemplate(resp, "base", exportData)
		if err != nil {
			ErrorPage(resp, http.StatusInternalServerError)
			log.Printf("writing template: %v", err)
			return
		}
	}
}
