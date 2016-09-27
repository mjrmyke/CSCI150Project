package main
import (
	"html/template"
	"net/http"
)

var tpl *template.Template

func init() {
	funcMap := template.FuncMap{
	
	}
	tpl = template.New("").Funcs(funcMap)
	tpl = template.Must(tpl.ParseGlob("public/templates/*.gohtml"))
	
}

func ServeTemplateWithParams(res http.ResponseWriter, templateName string, params interface{}) error {
	return tpl.ExecuteTemplate(res, templateName, &params)
}