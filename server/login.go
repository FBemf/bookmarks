package server

import (
	"fmt"
	"local/bookmarks/datastore"
	"local/bookmarks/templates"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type loginData struct {
	Message string
}

func loginPage(templates *templates.Templates, ds *datastore.Datastore) httprouter.Handle {
	return func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		req.ParseForm()
		failed := req.Form.Get("failed")
		var data loginData
		if failed != "" {
			data.Message = "Login failed"
		}
		_, valid, err := authenticateSession(ds, req)
		if err != nil {
			ErrorPage(resp, http.StatusInternalServerError)
			log.Printf("finding session: %s", err)
			return
		}
		if valid {
			http.Redirect(resp, req, "/", http.StatusFound)
		} else {
			err := templates.Login.ExecuteTemplate(resp, "base", data)
			if err != nil {
				ErrorPage(resp, http.StatusInternalServerError)
				log.Printf("writing template: %s", err)
				return
			}
			resp.Header().Set("Content-Type", "text/html; charset=UTF-8")
		}
	}
}

func doLogin(templates *templates.Templates, ds *datastore.Datastore) httprouter.Handle {
	return func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		err := req.ParseForm()
		if err != nil {
			ErrorPage(resp, http.StatusBadRequest)
			return
		}
		username := req.Form.Get("username")
		password := req.Form.Get("password")
		userId, allowed, err := ds.AuthenticateUser(username, password)
		if err != nil {
			ErrorPage(resp, http.StatusInternalServerError)
			log.Printf("authenticating user: %s", err)
			return
		}
		if allowed {
			cookie, err := ds.CreateSession(userId)
			if err != nil {
				ErrorPage(resp, http.StatusInternalServerError)
				log.Printf("creating session: %s", err)
				return
			}
			http.SetCookie(resp, &cookie)
			http.Redirect(resp, req, "/", http.StatusSeeOther)
		} else {
			http.Redirect(resp, req, "/login?failed=1", http.StatusSeeOther)
		}
	}
}

func logout(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
	http.SetCookie(resp, &http.Cookie{Name: datastore.AUTH_COOKIE_NAME})
	http.Redirect(resp, req, "/", http.StatusSeeOther)
}

func authenticateSession(ds *datastore.Datastore, req *http.Request) (string, bool, error) {
	cookies := req.Cookies()
	var sessionCookie string
	for _, cookie := range cookies {
		if cookie.Name == datastore.AUTH_COOKIE_NAME {
			sessionCookie = cookie.Value
			break
		}
	}
	if sessionCookie == "" {
		return "", false, nil
	}
	username, valid, err := ds.GetSession(sessionCookie)
	if err != nil {
		return "", false, fmt.Errorf("finding session: %w", err)
	}
	if !valid {
		return "", false, nil
	}
	return username, true, nil
}
