package server

import (
	"encoding/json"
	"io/ioutil"
	"local/bookmarks/datastore"
	"local/bookmarks/templates"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
)

const tokenType = "Bearer "

type keysData struct {
	Keys      []datastore.ApiKey
	CsrfToken string
}

func keys(templates *templates.Templates, ds *datastore.Datastore) sessionHandler {
	return func(session datastore.Session, resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		resp.Header().Set("Content-Type", "text/html; charset=UTF-8")
		keys, err := ds.ListKeys()
		if err != nil {
			ErrorPage(resp, http.StatusInternalServerError)
			log.Printf("retrieving keys: %s", err)
			return
		}

		err = templates.ApiKeys.ExecuteTemplate(resp, "base", keysData{keys, session.CsrfToken})
		if err != nil {
			ErrorPage(resp, http.StatusInternalServerError)
			log.Printf("writing template: %v", err)
			return
		}
	}
}

func createKey(templates *templates.Templates, ds *datastore.Datastore) sessionHandler {
	return func(session datastore.Session, resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		req.ParseForm()
		name := req.Form.Get("name")
		if name == "" {
			ErrorPage(resp, http.StatusBadRequest)
			return
		}
		err := ds.CreateKey(name)
		if err != nil {
			ErrorPage(resp, http.StatusInternalServerError)
			log.Printf("creating key: %s", err)
			return
		}

		http.Redirect(resp, req, keysPrefix, http.StatusSeeOther)
	}
}

func deleteKey(templates *templates.Templates, ds *datastore.Datastore) sessionHandler {
	return func(session datastore.Session, resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		key := params[0].Value
		id, err := strconv.Atoi(key)
		if err != nil {
			ErrorPage(resp, http.StatusBadRequest)
			return
		}
		err = ds.DeleteKey(int64(id))
		if err != nil {
			ErrorPage(resp, http.StatusInternalServerError)
			log.Printf("creating key: %s", err)
			return
		}

		http.Redirect(resp, req, keysPrefix, http.StatusSeeOther)
	}
}

type apiNewBookmarkData struct {
	Name        string   `json:"name"`
	Url         string   `json:"url"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

func corsOptions(allow []string) httprouter.Handle {
	return func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		allowString := strings.Join(allow, ", ")
		header := resp.Header()
		header.Set("Access-Control-Allow-Origin", "*")
		header.Set("Access-Control-Allow-Headers", "Authorization")
		header.Set("Access-Control-Allow-Methods", allowString)
		header.Set("Allow", allowString)
		resp.WriteHeader(http.StatusNoContent)
	}
}

func apiBookmarklet(ds *datastore.Datastore) httprouter.Handle {
	return func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		err := req.ParseForm()
		if err != nil {
			ErrorPage(resp, http.StatusBadRequest)
			return
		}
		authToken := req.Form.Get("auth")
		_, allowed, err := ds.CheckKey(authToken)
		if err != nil {
			resultJson(resp, http.StatusInternalServerError)
			log.Printf("authenticating api call: %s", err)
			return
		}
		if allowed {
			data := apiNewBookmarkData{
				Name:        req.Form.Get("name"),
				Url:         req.Form.Get("url"),
				Description: req.Form.Get("description"),
				Tags:        req.Form["tag"],
			}
			if data.Name == "" || data.Url == "" {
				ErrorPage(resp, http.StatusBadRequest)
				return
			}
			ds.CreateBookmark(data.Name, ensureProtocol(data.Url), data.Description, data.Tags)
			http.Redirect(resp, req, "/", http.StatusSeeOther)
		} else {
			resultJson(resp, http.StatusForbidden)
		}
	}
}

func apiNewBookmark(ds *datastore.Datastore) httprouter.Handle {
	return func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		header := resp.Header()
		header.Set("Access-Control-Allow-Origin", "*")
		header.Set("Access-Control-Allow-Headers", "Authorization")
		header.Set("Access-Control-Allow-Methods", req.Method)
		header.Set("Content-Type", "application/json; charset=UTF-8")
		authHeader := req.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, tokenType) {
			ErrorPage(resp, http.StatusBadRequest)
			return
		} else {
			authHeader = authHeader[len(tokenType):]
		}
		_, allowed, err := ds.CheckKey(authHeader)
		if err != nil {
			resultJson(resp, http.StatusInternalServerError)
			log.Printf("authenticating api call: %s", err)
			return
		}
		if allowed {
			jsonData, err := ioutil.ReadAll(req.Body)
			if err != nil {
				resultJson(resp, http.StatusBadRequest)
				return
			}
			var data apiNewBookmarkData
			err = json.Unmarshal(jsonData, &data)
			if err != nil {
				resultJson(resp, http.StatusBadRequest)
				return
			}
			if data.Name == "" || data.Url == "" {
				resultJson(resp, http.StatusBadRequest)
				return
			}
			ds.CreateBookmark(data.Name, ensureProtocol(data.Url), data.Description, data.Tags)
			resultJson(resp, http.StatusOK)
		} else {
			resultJson(resp, http.StatusForbidden)
		}
	}
}

func apiExport(ds *datastore.Datastore) httprouter.Handle {
	return func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		resp.Header().Set("Content-Type", "application/json; charset=UTF-8")
		authHeader := req.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, tokenType) {
			ErrorPage(resp, http.StatusBadRequest)
			return
		} else {
			authHeader = authHeader[len(tokenType):]
		}
		_, allowed, err := ds.CheckKey(authHeader)
		if err != nil {
			resultJson(resp, http.StatusInternalServerError)
			log.Printf("authenticating api call: %s", err)
			return
		}
		if allowed {
			exported, err := ds.Export()
			if err != nil {
				resultJson(resp, http.StatusInternalServerError)
				log.Printf("exporting data: %s", err)
				return
			}
			_, err = resp.Write(exported)
			if err != nil {
				log.Printf("writing response: %s", err)
			}
		} else {
			resultJson(resp, http.StatusForbidden)
		}
	}
}

type resultData struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func resultJson(resp http.ResponseWriter, code int) {
	resp.WriteHeader(code)
	resp.Header().Set("Content-Type", "text/json; charset=UTF-8")
	data, err := json.Marshal(resultData{code, http.StatusText(code)})
	if err != nil {
		log.Printf("marshaling json: %s", err)
	}
	_, err = resp.Write(data)
	if err != nil {
		log.Printf("writing response: %s", err)
	}
}
