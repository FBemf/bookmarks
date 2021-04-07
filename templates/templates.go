package templates

import (
	"html/template"
	"io/fs"
)

type Templates struct {
	Index        *template.Template
	EditBookmark *template.Template
	ViewBookmark *template.Template
}

func CreateTemplates(templateFS fs.FS) Templates {
	index := template.Must(template.ParseFS(templateFS, "pages/base.html", "pages/index.html"))
	edit := template.Must(template.ParseFS(templateFS, "pages/base.html", "pages/edit.html"))
	view := template.Must(template.ParseFS(templateFS, "pages/base.html", "pages/view.html"))
	return Templates{
		Index:        index,
		EditBookmark: edit,
		ViewBookmark: view,
	}
}
