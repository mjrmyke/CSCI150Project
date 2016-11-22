// Contains the handlers for our OAuth interactions.
package main

import (
	"net/http"

	"github.com/Esseh/goauth"
	"github.com/julienschmidt/httprouter"
)

func init() {
	goauth.GlobalSettings.ClientType = "appengine"
}

const (
	PATH_AUTH_OAUTH_GITHUB_Recieve  = "/login/github/oauth/recieve"
	PATH_AUTH_OAUTH_GITHUB_Send     = "/login/github/oauth/send"
	PATH_AUTH_OAUTH_DROPBOX_Send    = "/login/dropbox/oauth/send"
	PATH_AUTH_OAUTH_DROPBOX_Recieve = "/login/dropbox/oauth/recieve"
)

func INIT_OAUTH_Handlers(r *httprouter.Router) {
	r.GET(PATH_AUTH_OAUTH_GITHUB_Send, AUTH_OAUTH_GITHUB_Send)
	r.GET(PATH_AUTH_OAUTH_GITHUB_Recieve, AUTH_OAUTH_GITHUB_Recieve)
	r.GET(PATH_AUTH_OAUTH_DROPBOX_Send, AUTH_OAUTH_DROPBOX_Send)
	r.GET(PATH_AUTH_OAUTH_DROPBOX_Recieve, AUTH_OAUTH_DROPBOX_Recieve)
}

const (
	GITHUB_CLIENTID = "e0297346f88565c9f443"
	GITHUB_SECRETID = "7dd96d4a262a004aeffefe4b0af1a38e03b38d14"

	DROPBOX_Appkey    = "ddhu8e7nswl56yt"
	DROPBOX_Appsecret = "387kru0n9nb0qkk"
)

func AUTH_OAUTH_GITHUB_Send(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	var model goauth.GitHubToken
	goauth.Send(res, req, "http://csci150project.appspot.com/login/github/oauth/recieve", GITHUB_CLIENTID, &model)
}

func AUTH_OAUTH_GITHUB_Recieve(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	var token goauth.GitHubToken
	err := goauth.Recieve(res, req, "http://csci150project.appspot.com/login/github/oauth/recieve", GITHUB_CLIENTID, GITHUB_SECRETID, &token)
	if err != nil {
		http.Redirect(res, req, PATH_AUTH_Login+"/?ErrorResponse=Unable to Fetch Credentials at This Time", http.StatusSeeOther)
		return
	}
	email, err := token.Email(req)
	if err != nil {
		http.Redirect(res, req, PATH_AUTH_Login+"/?ErrorResponse=Unable to Fetch Credentials at This Time", http.StatusSeeOther)
		return
	}
	info, err := token.AccountInfo(req)
	if err != nil {
		http.Redirect(res, req, PATH_AUTH_Login+"/?ErrorResponse=Unable to Fetch Credentials at This Time", http.StatusSeeOther)
		return
	}
	OAuthLogin(req, res, email.Email, info.Login, "", token.State)
}

func AUTH_OAUTH_DROPBOX_Send(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	var model goauth.DropboxToken
	goauth.Send(res, req, "http://csci150project.appspot.com/login/dropbox/oauth/recieve", DROPBOX_Appkey, &model)
}

func AUTH_OAUTH_DROPBOX_Recieve(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	var token goauth.DropboxToken
	err := goauth.Recieve(res, req, "http://csci150project.appspot.com/login/dropbox/oauth/recieve", DROPBOX_Appkey, DROPBOX_Appsecret, &token)
	if err != nil {
		http.Redirect(res, req, PATH_AUTH_Login+"/?ErrorResponse=Unable to Fetch Credentials at This Time (1)", http.StatusSeeOther)
		return
	}
	info, err := token.AccountInfo(req)
	if err != nil {
		http.Redirect(res, req, PATH_AUTH_Login+"/?ErrorResponse=Unable to Fetch Credentials at This Time (2)", http.StatusSeeOther)
		return
	}
	OAuthLogin(req, res, info.Email, info.NameDetails.GivenName, info.NameDetails.Surname, token.State)
}
