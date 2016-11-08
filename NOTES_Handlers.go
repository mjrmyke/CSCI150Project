package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

const (
	PATH_NOTES_document = "/document"
)

func INIT_NOTES_HANDLERS(r *httprouter.Router) {
	r.GET(PATH_NOTES_document, NOTES_GET_DOCUMENT) // PATH_AUTH_Login 				= "/login"
	// r.POST(PATH_AUTH_Login, AUTH_POST_Login)                   //
}

func NOTES_GET_DOCUMENT(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	_, err := GetUserFromSession(req) // Check if a user is already logged in.
	if err != nil {
		http.Redirect(res, req, "/"+req.FormValue("redirect"), http.StatusSeeOther)
		return
	}

	ServeTemplateWithParams(res, "papersheet.gohtml", struct {
		HeaderData
		ErrorResponse, RedirectURL string
	}{
		HeaderData:    *MakeHeader(res, req, false, true),
		RedirectURL:   req.FormValue("redirect"),
		ErrorResponse: req.FormValue("ErrorResponse"),
	})
}
