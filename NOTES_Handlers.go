package main

import (
	"html/template"
	"net/http"
	"strconv"

	"github.com/Esseh/retrievable"
	"github.com/julienschmidt/httprouter"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

const (
	PATH_NOTES_document = "/document" //// TODO: remove prototype
	PATH_NOTES_New      = "/new"
	PATH_NOTES_View     = "/view/:ID"
	PATH_NOTES_Editor   = "/edit/:ID"
	PATH_NOTES_EditRaw  = "/rawedit/:ID"
)

func INIT_NOTES_HANDLERS(r *httprouter.Router) {
	r.GET(PATH_NOTES_New, NOTES_GET_New)
	r.POST(PATH_NOTES_New, NOTES_POST_New)
	r.GET(PATH_NOTES_View, NOTES_GET_View)
	r.GET(PATH_NOTES_Editor, NOTES_GET_Editor)
	r.POST(PATH_NOTES_Editor, NOTES_POST_Editor)
	r.GET(PATH_NOTES_EditRaw, NOTES_GET_EditRaw)
	r.POST(PATH_NOTES_EditRaw, NOTES_GET_EditRaw)
	r.GET(PATH_NOTES_document, NOTES_GET_DOCUMENT)   // TODO: remove prototype
	r.POST(PATH_NOTES_document, NOTES_POST_DOCUMENT) // TODO: remove prototype
}

func NOTES_GET_New(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	if MustLogin(res, req) {
		return
	}

	ServeTemplateWithParams(res, "newnote", struct {
		HeaderData
		ErrorResponse, RedirectURL string
	}{
		HeaderData:    *MakeHeader(res, req, false, true),
		RedirectURL:   req.FormValue("redirect"),
		ErrorResponse: req.FormValue("ErrorResponse"),
	})
}

func NOTES_POST_New(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	u, err := GetUserFromSession(req) // Check if a user is already logged in.
	ctx := appengine.NewContext(req)

	if err != nil {
		http.Redirect(res, req, "/"+req.FormValue("redirect"), http.StatusSeeOther)
		return
	}

	data := req.FormValue("note")
	title := req.FormValue("title")
	protected, boolerr := strconv.ParseBool(req.FormValue("protection"))
	if ErrorPage(ctx, res, nil, "Internal Server Error (2)", boolerr, http.StatusSeeOther) {
		return
	}

	NewContent := Content{
		Title:   title,
		Content: data,
	}

	key, err := retrievable.PlaceEntity(ctx, int64(0), &NewContent)
	if ErrorPage(ctx, res, nil, "Internal Server Error (2)", err, http.StatusSeeOther) {
		return
	}

	NewNote := Note{
		OwnerID:   int64(u.IntID),
		Protected: protected,
		ContentID: key.IntID(),
	}

	newkey, err := retrievable.PlaceEntity(ctx, int64(0), &NewNote)
	if ErrorPage(ctx, res, nil, "Internal Server Error (2)", err, http.StatusSeeOther) {
		return
	}
	log.Infof(ctx, "Information being submitted: ", NewNote, NewContent)
	http.Redirect(res, req, "/view/"+strconv.FormatInt(newkey.IntID(), 10), http.StatusSeeOther)
}

/// TODO: implement
func NOTES_GET_View(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	_, err := GetUserFromSession(req) // Check if a user is already logged in.
	ctx := appengine.NewContext(req)

	NoteKeyStr := params.ByName("ID")
	NoteKey, err := strconv.ParseInt(NoteKeyStr, 10, 64)
	if ErrorPage(ctx, res, nil, "Internal Server Error (2)", err, http.StatusSeeOther) {
		return
	}

	ViewNote := &Note{}
	ViewContent := &Content{}

	err = retrievable.GetEntity(ctx, NoteKey, ViewNote)
	if ErrorPage(ctx, res, nil, "Internal Server Error (2)", err, http.StatusSeeOther) {
		return
	}

	err = retrievable.GetEntity(ctx, ViewNote.ContentID, ViewContent)
	if ErrorPage(ctx, res, nil, "Internal Server Error (2)", err, http.StatusSeeOther) {
		return
	}

	Body := template.HTML(ViewContent.Content)

	ServeTemplateWithParams(res, "document", struct {
		HeaderData
		ErrorResponse, RedirectURL, Title string
		Content                           template.HTML
	}{
		HeaderData:    *MakeHeader(res, req, false, true),
		RedirectURL:   req.FormValue("redirect"),
		ErrorResponse: req.FormValue("ErrorResponse"),
		Title:         ViewContent.Title,
		Content:       Body,
	})

}

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
