package core

import (
	"bytes"
	"html/template"
	"net/http"
)

func ExecuteTemplate(w http.ResponseWriter, tpl template.Template, name string, data interface{}) {
	var buf bytes.Buffer
	err := tpl.ExecuteTemplate(&buf, name, data)
	Expect(err, "error executing template")
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	_, err = buf.WriteTo(w)
	Expect(err, "error writing template to response writer")
}
