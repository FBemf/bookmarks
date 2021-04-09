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

type UrlParams struct {
	Page   int
	Order  string
	Search string
}

func SetPage(page string, p UrlParams) (UrlParams, error) {
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

func SetOrder(order string, p UrlParams) (UrlParams, error) {
	var err error
	p.Order = order
	if p.Order != NORMAL_ORDER && order != REVERSE_ORDER {
		return p, fmt.Errorf("illegal order %s", p.Order)
	}
	return p, err
}

func SetSearch(search string, p UrlParams) (UrlParams, error) {
	var err error
	p.Search = search
	return p, err
}

func (p UrlParams) QueryString() template.URL {
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
	result := strings.Join(params, "&")
	if result != "" {
		result = "?" + result
	}
	return template.URL(result)
}

func DefaultUrlParams() UrlParams {
	return UrlParams{
		Page:  1,
		Order: NORMAL_ORDER,
	}
}

func GetUrlParams(req *http.Request) (UrlParams, error) {
	params := DefaultUrlParams()

	err := req.ParseForm()
	if err != nil {
		return UrlParams{}, fmt.Errorf("parsing request params: %w", err)
	}
	pageString := req.Form.Get("page")
	if pageString != "" {
		params.Page, err = strconv.Atoi(pageString)
		if err != nil || params.Page < 1 {
			return UrlParams{}, fmt.Errorf("parsing page: %w", err)
		}
	}
	order := req.Form.Get("order")
	if order != "" {
		if order != NORMAL_ORDER && order != REVERSE_ORDER {
			return UrlParams{}, fmt.Errorf("invalid order %s", order)
		}
		params.Order = order
	}
	params.Search = req.Form.Get("search")
	return params, nil
}
