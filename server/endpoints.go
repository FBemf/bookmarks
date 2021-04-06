package server

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)

type helloData struct {
	Time string
}

func helloWorld(templates *template.Template) httprouter.Handle {
	return func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		resp.Header().Set("Content-Type", "text/html; charset=UTF-8")
		err := templates.ExecuteTemplate(resp, "date.html",
			helloData{Time: time.Now().Format("3:04:05.00 PM MST Monday Jan 2 2006")})
		if err != nil {
			errorPage(resp, http.StatusInternalServerError)
			log.Printf("writing template: %v", err)
			return
		}
	}
}
func errorPage(resp http.ResponseWriter, code int) {
	http.Error(resp, fmt.Sprintf("%d %s", code, http.StatusText(code)), code)
}
