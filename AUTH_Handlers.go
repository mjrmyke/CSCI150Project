// Handlers dealing with authentication
package main

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Esseh/retrievable"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/appengine"
	"google.golang.org/appengine/mail"
)

const (
	PATH_AUTH_Login          = "/login"
	PATH_AUTH_ElevatedLogin  = "/elevatedlogin"
	PATH_AUTH_Logout         = "/logout"
	PATH_AUTH_Register       = "/register"
	PATH_AUTH_ForgotPassword = "/forgot"
	PATH_AUTH_ResetPassword  = "/reset"
)

func INIT_AUTH_HANDLERS(r *httprouter.Router) {
	r.GET(PATH_AUTH_Logout, AUTH_GET_Logout)                   // PATH_AUTH_Logout 				= "/logout"
	r.GET(PATH_AUTH_Login, AUTH_GET_Login)                     // PATH_AUTH_Login 				= "/login"
	r.POST(PATH_AUTH_Login, AUTH_POST_Login)                   //
	r.GET(PATH_AUTH_ElevatedLogin, AUTH_GET_ElevatedLogin)     // PATH_AUTH_Login 				= "/login"
	r.POST(PATH_AUTH_ElevatedLogin, AUTH_POST_ElevatedLogin)   //
	r.GET(PATH_AUTH_Register, AUTH_GET_Register)               // PATH_AUTH_Register 			= "/register"
	r.POST(PATH_AUTH_Register, AUTH_POST_Register)             //
	r.POST(PATH_AUTH_ForgotPassword, AUTH_POST_ForgotPassword) //
	r.GET(PATH_AUTH_ResetPassword, AUTH_GET_ResetPassword)     // PATH_AUTH_ResetPassword		= "/reset"
	r.POST(PATH_AUTH_ResetPassword, AUTH_POST_ResetPassword)   //
}

//=========================================================================================
// Login
//=========================================================================================
func AUTH_GET_Login(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	_, err := GetUserFromSession(req) // Check if a user is already logged in.
	if err == nil {
		http.Redirect(res, req, "/"+req.FormValue("redirect"), http.StatusSeeOther)
		return
	}

	ServeTemplateWithParams(res, "login", struct {
		HeaderData
		ErrorResponse, RedirectURL string
	}{
		HeaderData:    *MakeHeader(res, req, false, true),
		RedirectURL:   req.FormValue("redirect"),
		ErrorResponse: req.FormValue("ErrorResponse"),
	})
}

func AUTH_POST_Login(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	ctx := appengine.NewContext(req)
	username := strings.ToLower(req.FormValue("email"))
	password := req.FormValue("password")
	redirect := req.FormValue("redirect")

	if username == "" || password == "" { // Check incoming information for Trivial False case.
		v := url.Values{}
		v.Add("redirect", redirect)
		v.Add("ErrorResponse", "Fields Cannot Be Empty")
		http.Redirect(res, req, PATH_AUTH_Login+"?"+v.Encode(), http.StatusSeeOther)
		return
	}

	userID, err := GetUserIDFromLogin(ctx, username, password)
	if BackWithError(res, req, err, "Login Information Is Incorrect") {
		return
	}

	sessionID, err := CreateSessionID(ctx, req, userID)
	if BackWithError(res, req, err, "Login error, try again later.") {
		return
	}

	err = MakeCookie(res, "session", strconv.FormatInt(sessionID, 10))
	if BackWithError(res, req, err, "Login error, try again later.") {
		return
	}
	http.Redirect(res, req, "/"+redirect, http.StatusSeeOther)
}

//=========================================================================================
// ElevatedLogin
//=========================================================================================
func AUTH_GET_ElevatedLogin(res http.ResponseWriter, req *http.Request, params httprouter.Params) {

	ServeTemplateWithParams(res, "login", struct {
		HeaderData
		ErrorResponse, RedirectURL string
	}{
		HeaderData:    *MakeHeader(res, req, false, true),
		RedirectURL:   req.FormValue("redirect"),
		ErrorResponse: req.FormValue("ErrorResponse"),
	})
}

func AUTH_POST_ElevatedLogin(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	ctx := appengine.NewContext(req)
	username := strings.ToLower(req.FormValue("email"))
	password := req.FormValue("password")
	redirect := req.FormValue("redirect")

	if username == "" || password == "" { // Check incoming information for Trivial False case.
		v := url.Values{}
		v.Add("redirect", redirect)
		v.Add("ErrorResponse", "Fields Cannot Be Empty")
		http.Redirect(res, req, PATH_AUTH_Login+"?"+v.Encode(), http.StatusSeeOther)
		return
	}

	userID, err := GetUserIDFromLogin(ctx, username, password)
	if BackWithError(res, req, err, "Login Information Is Incorrect") {
		return
	}

	ElevatedPermStruct := ElevatedPerms{
		userid: userID,
	}

	err = retrievable.PlaceInMemcache(ctx, strconv.FormatInt(userID, 10), &ElevatedPermStruct, 60*time.Minute)
	if BackWithError(res, req, err, "Error saving login information") {
		return
	}

	http.Redirect(res, req, "/"+redirect, http.StatusSeeOther)
}

//=========================================================================================
//Logout
//=========================================================================================
func AUTH_GET_Logout(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	ctx := appengine.NewContext(req)
	sessionIDStr, err := GetCookieValue(req, "session")
	if ErrorPage(ctx, res, nil, "Must be logged in", err, http.StatusBadRequest) {
		return
	}

	sessionVal, err := strconv.ParseInt(sessionIDStr, 10, 0)
	if ErrorPage(ctx, res, nil, "Bad cookie value", err, http.StatusBadRequest) {
		return
	}

	err = retrievable.DeleteEntity(ctx, (&Session{}).Key(ctx, sessionVal))
	if ErrorPage(ctx, res, nil, "No such session found!", err, 500) {
		return
	}

	DeleteCookie(res, "session")
	http.Redirect(res, req, "/"+req.FormValue("redirect"), http.StatusSeeOther)
}

//=========================================================================================
//Register
//=========================================================================================
func AUTH_GET_Register(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	u, _ := GetUserFromSession(req) // Check if already logged in
	if u.IntID != 0 {
		http.Redirect(res, req, "/"+req.FormValue("redirect"), http.StatusSeeOther)
		return
	}
	ServeTemplateWithParams(res, "user-register", struct {
		HeaderData
		BusinessKey, ErrorResponse, RedirectURL string
	}{
		HeaderData:    *MakeHeader(res, req, true, false),
		ErrorResponse: req.FormValue("ErrorResponse"),
		BusinessKey:   req.FormValue("BusinessKey"),
		RedirectURL:   req.FormValue("redirect"),
	})
}
func AUTH_POST_Register(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	ctx := appengine.NewContext(req)
	nu := &User{ // Make the New User
		Email:    strings.ToLower(req.FormValue("email")),
		First:    req.FormValue("given-name"),
		Last:     req.FormValue("family-name"),
		Customer: "",
	}

	password := req.FormValue("password")
	confirmPassword := req.FormValue("cpassword")

	// Check for trivially false.
	if "" == nu.Email || "" == nu.First || "" == nu.Last || "" == password || "" == confirmPassword || password != confirmPassword {
		BackWithError(res, req, ErrEmptyField, "Field Empty or password mismatch")
		return
	}

	nu, err := CreateUserFromLogin(ctx, nu.Email, password, nu)
	if BackWithError(res, req, err, "Username Taken") {
		return
	}

	AUTH_POST_Login(res, req, params)
}

//=========================================================================================
//ForgotPassword
//=========================================================================================
func AUTH_POST_ForgotPassword(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	ctx := appengine.NewContext(req)
	uid, err := GetUserIDFromEmail(ctx, req.FormValue("email")) // Get the User ID as well as making sure the user actually exists.
	if err == nil {                                             // Proceed If the user is valid
		u := User{}
		err = retrievable.GetEntity(ctx, retrievable.IntID(uid), &u)
		if ErrorPage(ctx, res, nil, "Internal Server Error (1)", err, http.StatusSeeOther) {
			return
		}

		// Make one time password reset key
		key, retErr := retrievable.PlaceEntity(ctx, int64(0), &PasswordReset{UID: uid, Creation: time.Now()})
		if ErrorPage(ctx, res, nil, "Internal Server Error (2)", retErr, http.StatusSeeOther) {
			return
		}

		// Send Message
		msg := &mail.Message{
			Sender:  "noreply@greatercommons.com",
			To:      []string{req.FormValue("email")},
			Subject: "Greater Commons Forgotten Password",
			Body:    "Change your password at http://www." + req.URL.Host + ".com/change using the following code:\n" + strconv.FormatInt(key.IntID(), 10),
		}
		err = mail.Send(ctx, msg)
		if ErrorPage(ctx, res, nil, "Internal Server Error (2)", err, http.StatusSeeOther) {
			return
		}
	}

	http.Redirect(res, req, PATH_AUTH_ResetPassword, http.StatusSeeOther)
}

//=========================================================================================
//ResetPassword
//=========================================================================================
func AUTH_GET_ResetPassword(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	u, _ := GetUserFromSession(req)
	if u.IntID != 0 {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}
	ServeTemplateWithParams(res, "resetPassword", struct {
		HeaderData
		ErrorResponse string
	}{
		HeaderData:    *MakeHeader(res, req, true, true),
		ErrorResponse: req.FormValue("ErrorResponse"),
	})
}
func AUTH_POST_ResetPassword(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	key := req.FormValue("key")
	password := req.FormValue("password")
	confirmPassword := req.FormValue("confirmPassword")

	if password != confirmPassword || len(password) < 4 || key == "" {
		http.Redirect(res, req, PATH_AUTH_ResetPassword, http.StatusSeeOther)
		return
	}

	ctx := appengine.NewContext(req)

	// Retrieve the Password Reset
	pr := PasswordReset{}
	prKey, _ := strconv.ParseInt(key, 10, 64)
	getErr := retrievable.GetEntity(ctx, prKey, &pr)
	if ErrorPage(ctx, res, nil, "Internal Server Error (1)", getErr, http.StatusSeeOther) {
		return
	}

	// Retrieve the User
	u := User{}
	getErr2 := retrievable.GetEntity(ctx, pr.UID, &u)
	if ErrorPage(ctx, res, nil, "Internal Server Error (2)", getErr2, http.StatusSeeOther) {
		return
	}

	// Generate New Password
	cryptPass, cryptErr := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if ErrorPage(ctx, res, nil, "Internal Server Error (3)", cryptErr, http.StatusSeeOther) {
		return
	}

	// Get the Login Information
	loginAccount := LoginLocalAccount{}
	getErr3 := retrievable.GetEntity(ctx, u.Email, &loginAccount)
	if ErrorPage(ctx, res, nil, "Internal Server Error (4)", getErr3, http.StatusSeeOther) {
		return
	}

	// Update the Login Information
	loginAccount.Password = cryptPass
	_, placeErr := retrievable.PlaceEntity(ctx, u.Email, &loginAccount)
	if ErrorPage(ctx, res, nil, "Internal Server Error (4)", placeErr, http.StatusSeeOther) {
		return
	}

	// Get the datastore key of the password reset entry
	cleanup, placeErr2 := retrievable.PlaceEntity(ctx, prKey, &pr)
	if ErrorPage(ctx, res, nil, "Internal Server Error (5)", placeErr2, http.StatusSeeOther) {
		return
	}

	// Delete the password reset entry
	delErr := retrievable.DeleteEntity(ctx, cleanup)
	if ErrorPage(ctx, res, nil, "Internal Server Error (6)", delErr, http.StatusSeeOther) {
		return
	}

	http.Redirect(res, req, PATH_AUTH_Login, http.StatusSeeOther)
}
