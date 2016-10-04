package main

import (
	"html/template"
	"net/http"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
)

// Global Template file.
var tpl *template.Template

func init() {
	// Tie functions into template here with ... "functionName":theFunction,
	funcMap := template.FuncMap{}
	// Load up all templates.
	tpl = template.New("").Funcs(funcMap)
	tpl = template.Must(tpl.ParseGlob("public/templates/*.gohtml"))

}

func ServeTemplateWithParams(res http.ResponseWriter, templateName string, params interface{}) error {
	return tpl.ExecuteTemplate(res, templateName, &params)
}

// Header Data,
// Present in most template executions. (Unless it's an internal it should be assumed to be used.)
type HeaderData struct {
	ShowLogin    bool
	ShowRegister bool
	Ctx          context.Context
	User         *User
	CurrentPath  string
}

// Constructs the header.
// As the header gets more complex(such as capturing the current path)
// the need for such a helper function increases.
func MakeHeader(res http.ResponseWriter, req *http.Request, login, register bool) *HeaderData {
	u, _ := GetUserFromSession(req)
	oldCookie, err := GetCookieValue(req, "session")
	if err == nil {
		MakeCookie(res, "session", oldCookie)
	}
	redirectURL := req.URL.Path[1:]
	if redirectURL == "login" || redirectURL == "register" || redirectURL == "elevatedlogin" {
		redirectURL = req.URL.Query().Get("redirect")
	}
	return &HeaderData{
		login, register, appengine.NewContext(req), u, redirectURL,
	}
}
