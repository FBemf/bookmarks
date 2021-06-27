package server

import (
	"fmt"
	"local/bookmarks/datastore"
	"local/bookmarks/templates"
	"log"
	"net/http"
	"net/url"

	"github.com/julienschmidt/httprouter"
)

type loginData struct {
	Message    string
	RedirectTo string
}

func loginPage(templates *templates.Templates, ds *datastore.Datastore) httprouter.Handle {
	return func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		resp.Header().Set("Content-Type", "text/html; charset=UTF-8")
		req.ParseForm()
		failed := req.Form.Get("failed")
		redirectTo := req.Form.Get("redirectTo")
		var data loginData
		if failed != "" {
			data.Message = "Login failed"
		}
		data.RedirectTo = redirectTo
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
		}
	}
}

func doLogin(templates *templates.Templates, ds *datastore.Datastore) httprouter.Handle {
	tryAgainUrl, err := url.Parse("/login?failed=1")
	if err != nil {
		log.Panicf("tried to parse a bad url path: %s", err)
	}
	return func(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
		err := req.ParseForm()
		if err != nil {
			ErrorPage(resp, http.StatusBadRequest)
			return
		}
		username := req.Form.Get("username")
		password := req.Form.Get("password")
		redirectTo := req.Form.Get("redirectTo")
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
			log.Printf("redirecting to %s", redirectTo)
			http.Redirect(resp, req, redirectTo, http.StatusSeeOther)
		} else {
			redirectUrl := tryAgainUrl
			q := redirectUrl.Query()
			q.Set("redirectTo", redirectTo)
			redirectUrl.RawQuery = q.Encode()
			http.Redirect(resp, req, redirectUrl.String(), http.StatusSeeOther)
		}
	}
}

func logout(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
	http.SetCookie(resp, &http.Cookie{Name: datastore.AuthCookieName})
	http.Redirect(resp, req, "/", http.StatusSeeOther)
}

func authenticateSession(ds *datastore.Datastore, req *http.Request) (datastore.Session, bool, error) {
	cookies := req.Cookies()
	var sessionCookie string
	for _, cookie := range cookies {
		if cookie.Name == datastore.AuthCookieName {
			sessionCookie = cookie.Value
			break
		}
	}
	if sessionCookie == "" {
		return datastore.Session{}, false, nil
	}
	session, valid, err := ds.GetSession(sessionCookie)
	if err != nil {
		return datastore.Session{}, false, fmt.Errorf("finding session: %w", err)
	}
	if !valid {
		return datastore.Session{}, false, nil
	}
	return session, true, nil
}
