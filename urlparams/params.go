package urlparams

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

type UrlParams struct {
	Page int
}

func SetPage(page string, p UrlParams) (UrlParams, error) {
	var err error
	p.Page, err = strconv.Atoi(page)
	return p, err
}

func (p UrlParams) QueryString() template.URL {
	params := make([]string, 0)
	if p.Page != 1 {
		params = append(params, "page="+strconv.Itoa(p.Page))
	}
	result := strings.Join(params, "&")
	if result != "" {
		result = "?" + result
	}
	return template.URL(result)
}

func GetUrlParams(req *http.Request) (UrlParams, error) {
	err := req.ParseForm()
	if err != nil {
		return UrlParams{}, fmt.Errorf("parsing request params: %w", err)
	}
	pageString := req.Form.Get("page")
	var page int
	if pageString == "" {
		page = 1
	} else {
		page, err = strconv.Atoi(pageString)
		if err != nil || page < 1 {
			return UrlParams{}, fmt.Errorf("parsing page: %w", err)
		}
	}
	return UrlParams{page}, nil
}
