package templates

import (
	"html/template"
	"io/fs"
	"local/bookmarks/urlparams"
)

type Templates struct {
	Login        *template.Template
	Index        *template.Template
	EditBookmark *template.Template
	ViewBookmark *template.Template
}

func functions() *template.Template {
	return template.New("").
		Funcs(template.FuncMap{
			"paramSetPage":     urlparams.SetPage,
			"paramSetOrder":    urlparams.SetOrder,
			"paramSetSearch":   urlparams.SetSearch,
			"paramClearTags":   urlparams.ClearTags,
			"paramQueryString": urlparams.SearchParams.QueryString,
		})
}

func CreateTemplates(templateFS fs.FS) Templates {
	login := template.Must(functions().ParseFS(templateFS, "pages/base.html", "pages/login.html"))
	index := template.Must(functions().ParseFS(templateFS, "pages/base.html", "pages/index.html"))
	edit := template.Must(functions().ParseFS(templateFS, "pages/base.html", "pages/edit.html"))
	view := template.Must(functions().ParseFS(templateFS, "pages/base.html", "pages/view.html"))
	return Templates{
		Login:        login,
		Index:        index,
		EditBookmark: edit,
		ViewBookmark: view,
	}
}
