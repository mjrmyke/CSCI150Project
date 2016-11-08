// The USERS Module, Deals with the User interfacing with themselves.
package main

import (
	"net/http"
	"strconv"
	"time"

	"google.golang.org/appengine/log"

	"github.com/Esseh/retrievable"
	"github.com/julienschmidt/httprouter"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

func INIT_USERS_HANDLERS(r *httprouter.Router) {
	r.POST(PATH_USERS_ChangeInfo, USERS_POST_ChangeInfo)
	r.POST(PATH_USERS_DeleteAccount, USERS_POST_DeleteAccount)
	r.GET(PATH_USERS_GET_Profile, USERS_GET_Profile)         // PATH_USERS_GET_Profile 	  = "/profile"
	r.GET(PATH_USERS_GET_ProfileView, USERS_GET_ProfileView) // PATH_USERS_GET_ProfileView 	  = "/profile/:ID"
	r.GET(PATH_USERS_Sessions, USERS_GET_Sessions)
	r.GET(PATH_USERS_DeleteSession, USERS_GET_DeleteSession)
	r.GET(PATH_USERS_DeleteAllSessions, USERS_GET_DeleteAllSessions)
	r.GET(PATH_USERS_Terms, USERS_GET_Terms)
	r.POST(PATH_USERS_Changepassword, USERS_POST_Changepassword)
	r.POST(PATH_USERS_Changeavatar, USERS_POST_Changeavatar)
	r.POST(PATH_USERS_Changeemail, USERS_POST_Changeemail)
}

const (
	PATH_USERS_GET_Profile       = "/profile"
	PATH_USERS_GET_ProfileView   = "/profile/:ID"
	PATH_USERS_ChangeInfo        = "/changeinfo"
	PATH_USERS_DeleteAccount     = "/deleteaccount"
	PATH_USERS_Sessions          = "/sessions"
	PATH_USERS_DeleteSession     = "/sessions/delete/:ID"
	PATH_USERS_DeleteAllSessions = "/sessions/delete"
	PATH_USERS_Terms             = "/terms"
	PATH_USERS_Changepassword    = "/api/profile/changepassword"
	PATH_USERS_Changeavatar      = "/api/profile/changeavatar"
	PATH_USERS_Changeemail       = "/api/profile/changeemail"
)

//===========================================================================
// Get Terms
//===========================================================================
func USERS_GET_Terms(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	ServeTemplateWithParams(res, "user-legal-terms.gohtml", struct {
		HeaderData
	}{
		HeaderData: *MakeHeader(res, req, true, true),
	})
}

//===========================================================================
// Delete All Sessions
//===========================================================================
func USERS_GET_DeleteAllSessions(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	if MustLogin(res, req) {
		return
	}
	ctx := appengine.NewContext(req)
	currSesID, err := GetSessionID(req)
	if ErrorPage(ctx, res, nil, "Unable to get session ID from cookie", err, http.StatusBadRequest) {
		return
	}
	currSes, err := GetSession(ctx, currSesID)
	if ErrorPage(ctx, res, nil, "Unable to get session from session ID", err, http.StatusBadRequest) {
		return
	}

	err = DeleteAllOtherSessionsForUser(ctx, currSes.UserID, currSesID)
	if ErrorPage(ctx, res, nil, "Unable to delete sessions", err, http.StatusInternalServerError) {
		return
	}
	http.Redirect(res, req, "/sessions", http.StatusSeeOther)
}

//===========================================================================
// Delete Session
//===========================================================================
func USERS_GET_DeleteSession(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	if MustLogin(res, req) {
		return
	}
	ctx := appengine.NewContext(req)
	u, err := GetUserFromSession(req)
	if ErrorPage(ctx, res, nil, "Unable to get user account", err, http.StatusInternalServerError) {
		return
	}
	id, err := strconv.ParseInt(params.ByName("ID"), 10, 64)
	if ErrorPage(ctx, res, u, "Session ID must be an integer", err, http.StatusBadRequest) {
		return
	}
	err = DeleteUserSession(ctx, int64(u.IntID), id)
	if err == ErrNoSession {
		ErrorPage(ctx, res, u, "Session does not exist", err, http.StatusInternalServerError)
		return
	} else if ErrorPage(ctx, res, u, "Unable to delete session", err, http.StatusInternalServerError) {
		return
	}
	http.Redirect(res, req, "/sessions", http.StatusSeeOther)
}

//===========================================================================
// Sessions
//===========================================================================
func USERS_GET_Sessions(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	if MustLogin(res, req) {
		return
	}
	ctx := appengine.NewContext(req)
	currSesID, err := GetSessionID(req)
	if ErrorPage(ctx, res, nil, "Unable to get session ID from cookie", err, http.StatusBadRequest) {
		return
	}
	currSes, err := GetSession(ctx, currSesID)
	if ErrorPage(ctx, res, nil, "Unable to get session from session ID", err, http.StatusBadRequest) {
		return
	}
	ss, err := GetAllSessionsForUser(ctx, currSes.UserID)
	if ErrorPage(ctx, res, nil, "Error getting user sessions", err, http.StatusInternalServerError) {
		return
	}

	for i, v := range ss {
		if v.ID == currSes.ID {
			ss = append(ss[:i], ss[i+1:]...)
			break
		}
	}

	ServeTemplateWithParams(res, "user-sessions.gohtml", struct {
		HeaderData
		CurrentSession Session
		OtherSessions  []Session
	}{
		HeaderData:     *MakeHeader(res, req, true, true),
		CurrentSession: currSes,
		OtherSessions:  ss,
	})
}

//===========================================================================
// Profile
//===========================================================================
func USERS_GET_Profile(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	if MustLogin(res, req) {
		return
	}
	ctx := appengine.NewContext(req)
	log.Infof(ctx, "Request query string is: %s", req.URL.Query())
	ServeTemplateWithParams(res, "profile.gohtml", struct {
		HeaderData
		ErrorResponseProfile string
	}{
		*MakeHeader(res, req, true, true),
		req.FormValue("ErrorResponseProfile"),
	})
}

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
	ServeTemplateWithParams(res, "user-profile.gohtml", screen)
}

//===========================================================================
// Change Info
//===========================================================================
func USERS_POST_ChangeInfo(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	u, err := GetUserFromSession(req)
	if DestinationWithError(res, req, PATH_USERS_GET_Profile+"#", err, "Not Logged In") {
		return
	}
	u.First = req.FormValue("firstName")
	u.Last = req.FormValue("lastName")
	// TODO: Make sure we can parse any date HTML element input.
	newBirthday := req.FormValue("birthday")
	if newBirthday != "" {
		u.DOB, err = time.Parse("1/2/2006", newBirthday)
		if DestinationWithErrorAt(res, req, PATH_USERS_GET_Profile+"#", err, "ErrorResponseProfile", "Bad Birthday Input") {
			return
		}
	}
	u.Bio = req.FormValue("user-bio")

	ctx := appengine.NewContext(req)
	log.Infof(ctx, "Bio is: %s\n", req.FormValue("user-bio"))
	_, placeErr := retrievable.PlaceEntity(ctx, u.IntID, u)
	if DestinationWithError(res, req, PATH_USERS_GET_Profile+"#", placeErr, "Internal Server Error") {
		return
	}
	http.Redirect(res, req, PATH_USERS_GET_Profile+"#", http.StatusSeeOther)
}

//===========================================================================
// Delete Account
//===========================================================================
func USERS_POST_DeleteAccount(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	u, err := GetUserFromSession(req)
	if DestinationWithError(res, req, PATH_USERS_GET_Profile, err, "Not Logged In") {
		return
	}
	ctx := appengine.NewContext(req)

	err = ReqElevatedUserPerms(ctx, u.IntID, res, req)
	if err != nil {
		return
	}

	uid, err := GetUserIDFromLogin(ctx, req.FormValue("username"), req.FormValue("password"))
	if DestinationWithError(res, req, PATH_USERS_GET_Profile, err, "Bad Credentials") {
		return
	}
	if int64(u.IntID) != uid {
		DestinationWithError(res, req, PATH_USERS_GET_Profile, ErrPasswordMatch, "Not Logged In")
		return
	}
	err = retrievable.DeleteEntity(ctx, datastore.NewKey(ctx, LoginTable, req.FormValue("username"), 0, nil))
	if DestinationWithError(res, req, PATH_USERS_GET_Profile, err, "Internal Server Error") {
		return
	}
	err = retrievable.DeleteEntity(ctx, u.Key(ctx, u.IntID))
	if DestinationWithError(res, req, PATH_USERS_GET_Profile, err, "Internal Server Error") {
		return
	}
	DeleteCookie(res, "session")
	http.Redirect(res, req, "/", http.StatusSeeOther)
}

//===========================================================================
// Change Password
//===========================================================================
func USERS_POST_Changepassword(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	if MustLogin(res, req) {
		return
	}
	u, _ := GetUserFromSession(req)
	ctx := appengine.NewContext(req)
	oldPass := req.FormValue("old-pass")
	newPass := req.FormValue("new-pass")
	newPassConfirm := req.FormValue("new-pass-confirm")
	if newPass != newPassConfirm {
		DestinationWithError(res, req, PATH_USERS_GET_Profile+"#password", ErrPasswordMatch, "Passwords Do Not Match")
		return
	}
	err := ChangePassword(ctx, u.Email, oldPass, newPass)
	if DestinationWithError(res, req, PATH_USERS_GET_Profile+"#password", err, "Problem Changing Password, Try Again Later") {
		return
	}
	http.Redirect(res, req, PATH_USERS_GET_Profile+"#password", http.StatusSeeOther)
}

//===========================================================================
// Change Avatar
//===========================================================================
func USERS_POST_Changeavatar(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	if MustLogin(res, req) {
		return
	}
	u, _ := GetUserFromSession(req)
	ctx := appengine.NewContext(req)
	rdr, hdr, err := req.FormFile("avatar")
	defer rdr.Close()
	if DestinationWithError(res, req, PATH_USERS_GET_Profile+"#", err, "Problem Recieving File") {
		return
	}
	err = UploadAvatar(ctx, int64(u.IntID), hdr, rdr)
	if DestinationWithError(res, req, PATH_USERS_GET_Profile+"#", err, "Internal Server Error") {
		return
	}
	u.Avatar = true
	_, err = retrievable.PlaceEntity(ctx, u.IntID, u)
	if DestinationWithError(res, req, PATH_USERS_GET_Profile+"#", err, "Internal Server Error") {
		return
	}
	http.Redirect(res, req, PATH_USERS_GET_Profile+"#", http.StatusSeeOther)
}

//===========================================================================
// Change Email
//===========================================================================
func USERS_POST_Changeemail(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	if MustLogin(res, req) {
		return
	}
	u, _ := GetUserFromSession(req)
	ctx := appengine.NewContext(req)
	if req.FormValue("email") == "" {
		DestinationWithError(res, req, PATH_USERS_GET_Profile+"#email", ErrEmptyField, "Email Cannot be Empty")
		return
	}
	oldEmail := u.Email
	u.Email = req.FormValue("email")
	err := ChangeEmail(ctx, oldEmail, req.FormValue("email"))
	if DestinationWithError(res, req, PATH_USERS_GET_Profile+"#email", err, "Cannot Change Email At This Time, Try Again Later") {
		return
	}
	_, err = retrievable.PlaceEntity(ctx, u.IntID, u)
	if DestinationWithError(res, req, PATH_USERS_GET_Profile+"#email", err, "Internal Server Error") {
		return
	}
	http.Redirect(res, req, PATH_USERS_GET_Profile+"#email", http.StatusSeeOther)
}
