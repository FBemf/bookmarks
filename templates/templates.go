package templates

import (
	"html/template"
	"io/fs"
	"local/bookmarks/urlparams"
)

type Templates struct {
	Index        *template.Template
	EditBookmark *template.Template
	ViewBookmark *template.Template
}

func functions() *template.Template {
	return template.New("").
		Funcs(template.FuncMap{
			"paramSetPage":     urlparams.SetPage,
			"paramQueryString": urlparams.UrlParams.QueryString,
		})
}

func CreateTemplates(templateFS fs.FS) Templates {
	index := template.Must(functions().ParseFS(templateFS, "pages/base.html", "pages/index.html"))
	edit := template.Must(functions().ParseFS(templateFS, "pages/base.html", "pages/edit.html"))
	view := template.Must(functions().ParseFS(templateFS, "pages/base.html", "pages/view.html"))
	return Templates{
		Index:        index,
		EditBookmark: edit,
		ViewBookmark: view,
	}
}
