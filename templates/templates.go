package templates

import (
	"fmt"
	"html/template"
	"io/fs"
	"local/bookmarks/datastore"
	"local/bookmarks/urlparams"
)

const CsrfTokenName = "csrf-token"

type Templates struct {
	Login        *template.Template
	ApiKeys      *template.Template
	Export       *template.Template
	Tags         *template.Template
	Index        *template.Template
	EditBookmark *template.Template
	ViewBookmark *template.Template
}

func functions() *template.Template {
	return template.New("").
		Funcs(template.FuncMap{
			"bookmarkAndParams": bookmarkAndParams,
			"emptyBookmark":     emptyBookmark,
			"paramSetPage":      urlparams.SetPage,
			"paramSetOrder":     urlparams.SetOrder,
			"paramSetSearch":    urlparams.SetSearch,
			"paramClearTags":    urlparams.ClearTags,
			"paramAddTag":       urlparams.AddTag,
			"paramQueryString":  urlparams.SearchParams.QueryString,
			"csrfField":         csrfField,
		})
}

func CreateTemplates(templateFS fs.FS) Templates {
	login := template.Must(functions().ParseFS(templateFS, "pages/base.html", "pages/login.html"))
	apiKeys := template.Must(functions().ParseFS(templateFS, "pages/base.html", "pages/keys.html"))
	export := template.Must(functions().ParseFS(templateFS, "pages/base.html", "pages/export.html"))
	index := template.Must(functions().ParseFS(templateFS, "pages/base.html", "pages/index.html"))
	tags := template.Must(functions().ParseFS(templateFS, "pages/base.html", "pages/tags.html"))
	edit := template.Must(functions().ParseFS(templateFS, "pages/base.html", "pages/edit.html"))
	view := template.Must(functions().ParseFS(templateFS, "pages/base.html", "pages/view.html"))
	return Templates{
		Login:        login,
		ApiKeys:      apiKeys,
		Export:       export,
		Index:        index,
		Tags:         tags,
		EditBookmark: edit,
		ViewBookmark: view,
	}
}

type bookmarkAndParamsData struct {
	Bookmark     datastore.Bookmark
	SearchParams urlparams.SearchParams
}

func bookmarkAndParams(bookmark datastore.Bookmark, params urlparams.SearchParams) bookmarkAndParamsData {
	return bookmarkAndParamsData{bookmark, params}
}

func emptyBookmark() datastore.Bookmark {
	return datastore.Bookmark{}
}

func csrfField(token string) template.HTML {
	return template.HTML(fmt.Sprintf("<input name=\"%s\" type=hidden value=\"%s\">", CsrfTokenName, token))
}
