// The USERS Module, Deals with the User interfacing with themselves.
package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"google.golang.org/appengine"
)

func INIT_USERS_HANDLERS(r *httprouter.Router) {
	r.GET(PATH_USERS_ProfileEdit, USERS_GET_ProfileEdit)
	r.POST(PATH_USERS_ProfileEdit, USERS_POST_ProfileEdit)
	r.GET(PATH_USERS_ProfileView, USERS_GET_ProfileView)
}

const (
	PATH_USERS_ProfileEdit = "/editprofile"
	PATH_USERS_ProfileView = "/profile/:ID"
)

//===========================================================================
// Profile
//===========================================================================
func USERS_GET_ProfileEdit(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	u, _ := GetUserFromSession(req)
	err := ServeTemplateWithParams(res, "profile-settings", struct {
		HeaderData
		ErrorResponseProfile string
		User *User
	}{
		*MakeHeader(res, req, true, true),
		req.FormValue("ErrorResponseProfile"),
		u,
	})
	if err != nil {
		fmt.Fprint(res,err)
	}
}

// TODO: Implement
func USERS_POST_ProfileEdit(res http.ResponseWriter, req *http.Request, params httprouter.Params) {}

//===========================================================================
// Profile View
//===========================================================================
func USERS_GET_ProfileView(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	ctx := appengine.NewContext(req)
	u, _ := GetUserFromSession(req)
	id, convErr := strconv.ParseInt(params.ByName("ID"), 10, 64)
	if ErrorPage(ctx, res, nil, "Invalid ID", convErr, http.StatusBadRequest) {
		return
	}
	ci, getErr := GetUserFromID(ctx, id)
	if ErrorPage(ctx, res, u, "Not a valid user ID", getErr, http.StatusNotFound) {
		return
	}
	screen := struct {
		HeaderData
		Data *User
	}{
		*MakeHeader(res, req, true, true),
		ci,
	}
	ServeTemplateWithParams(res, "user-profile", screen)
}
