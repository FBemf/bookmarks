package urlparams

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

const (
	REVERSE_ORDER = "reverse"
	NORMAL_ORDER  = "normal"
)

type SearchParams struct {
	Page       int
	Order      string
	Search     string
	SearchTags []string
}

func SetPage(page string, p SearchParams) (SearchParams, error) {
	var err error
	p.Page, err = strconv.Atoi(page)
	if err != nil {
		return p, err
	}
	if p.Page < 1 {
		return p, fmt.Errorf("illegal page %d", p.Page)
	}
	return p, err
}

func SetOrder(order string, p SearchParams) (SearchParams, error) {
	var err error
	p.Order = order
	if p.Order != NORMAL_ORDER && order != REVERSE_ORDER {
		return p, fmt.Errorf("illegal order %s", p.Order)
	}
	return p, err
}

func SetSearch(search string, p SearchParams) (SearchParams, error) {
	var err error
	p.Search = search
	return p, err
}

func ClearTags(p SearchParams) (SearchParams, error) {
	var err error
	p.SearchTags = make([]string, 0)
	return p, err
}

func (p SearchParams) QueryString() template.URL {
	params := make([]string, 0)
	if p.Page != 1 {
		params = append(params, "page="+strconv.Itoa(p.Page))
	}
	if p.Order != NORMAL_ORDER {
		params = append(params, "order="+p.Order)
	}
	if p.Search != "" {
		params = append(params, "search="+p.Search)
	}
	for _, tag := range p.SearchTags {
		params = append(params, "searchTag="+tag)
	}
	result := strings.Join(params, "&")
	if result != "" {
		result = "?" + result
	}
	return template.URL(result)
}

func DefaultUrlParams() SearchParams {
	return SearchParams{
		Page:       1,
		Order:      NORMAL_ORDER,
		Search:     "",
		SearchTags: []string{},
	}
}

func GetUrlParams(req *http.Request) (SearchParams, error) {
	params := DefaultUrlParams()

	err := req.ParseForm()
	if err != nil {
		return SearchParams{}, fmt.Errorf("parsing request params: %w", err)
	}
	pageString := req.Form.Get("page")
	if pageString != "" {
		params.Page, err = strconv.Atoi(pageString)
		if err != nil || params.Page < 1 {
			return SearchParams{}, fmt.Errorf("parsing page: %w", err)
		}
	}
	order := req.Form.Get("order")
	if order != "" {
		if order != NORMAL_ORDER && order != REVERSE_ORDER {
			return SearchParams{}, fmt.Errorf("invalid order %s", order)
		}
		params.Order = order
	}
	params.Search = req.Form.Get("search")
	params.SearchTags = make([]string, 0, len(req.Form["searchTag"]))
	params.SearchTags = append(params.SearchTags, req.Form["searchTag"]...)
	return params, nil
}
