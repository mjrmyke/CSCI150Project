package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

const (
	PATH_NOTES_document = "/document" //// TODO: remove prototype
	PATH_NOTES_New      = "/new"
	PATH_NOTES_View     = "/view/:ID/:TITLE"
	PATH_NOTES_Editor   = "/edit/:ID"
	PATH_NOTES_EditRaw  = "/rawedit/:ID"
)

func INIT_NOTES_HANDLERS(r *httprouter.Router) {
	r.GET(PATH_NOTES_New, NOTES_GET_New)
	r.GET(PATH_NOTES_View, NOTES_GET_View)
	r.GET(PATH_NOTES_Editor, NOTES_GET_Editor)
	r.POST(PATH_NOTES_Editor, NOTES_POST_Editor)
	r.GET(PATH_NOTES_EditRaw, NOTES_GET_EditRaw)
	r.POST(PATH_NOTES_EditRaw, NOTES_GET_EditRaw)
	r.GET(PATH_NOTES_document, NOTES_GET_DOCUMENT)   // TODO: remove prototype
	r.POST(PATH_NOTES_document, NOTES_POST_DOCUMENT) // TODO: remove prototype
}

/// TODO: implement
func NOTES_GET_New(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	_, err := GetUserFromSession(req) // Check if a user is already logged in.
	if err != nil {
		http.Redirect(res, req, "/"+req.FormValue("redirect"), http.StatusSeeOther)
		return
	}
	ctx := appengine.NewContext(req)

	id := params.ByName("ID")

	title := params.ByName("TITLE")

	key := id + "|" + title

	log.Infof(ctx, "im just here so i allow you to compile", key)
	ServeTemplateWithParams(res, "document.gohtml", struct {
		HeaderData
		ErrorResponse, RedirectURL string
	}{
		HeaderData:    *MakeHeader(res, req, false, true),
		RedirectURL:   req.FormValue("redirect"),
		ErrorResponse: req.FormValue("ErrorResponse"),
	})
}

/// TODO: implement
func NOTES_GET_View(res http.ResponseWriter, req *http.Request, params httprouter.Params) {}

/// TODO: implement
func NOTES_GET_Editor(res http.ResponseWriter, req *http.Request, params httprouter.Params) {}

/// TODO: implement
func NOTES_POST_Editor(res http.ResponseWriter, req *http.Request, params httprouter.Params) {}

/// TODO: implement
func NOTES_GET_EditRaw(res http.ResponseWriter, req *http.Request, params httprouter.Params) {}

/// TODO: implement
func NOTES_POST_EditRaw(res http.ResponseWriter, req *http.Request, params httprouter.Params) {}

//// TODO: remove prototype
func NOTES_GET_DOCUMENT(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	_, err := GetUserFromSession(req) // Check if a user is already logged in.
	if err != nil {
		http.Redirect(res, req, "/"+req.FormValue("redirect"), http.StatusSeeOther)
		return
	}

	ServeTemplateWithParams(res, "document", struct {
		HeaderData
		ErrorResponse, RedirectURL string
	}{
		HeaderData:    *MakeHeader(res, req, false, true),
		RedirectURL:   req.FormValue("redirect"),
		ErrorResponse: req.FormValue("ErrorResponse"),
	})
}

func NOTES_POST_DOCUMENT(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	_, err := GetUserFromSession(req) // Check if a user is already logged in.
	ctx := appengine.NewContext(req)

	if err != nil {
		http.Redirect(res, req, "/"+req.FormValue("redirect"), http.StatusSeeOther)
		return
	}
	data := req.FormValue("note")
	title := req.FormValue("title")
	log.Infof(ctx, "info from js:", title, data)
	http.Redirect(res, req, "/", http.StatusSeeOther)
}
