package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Reference Multiplexer functions in other modules here.
// Those functions splinter off the handlers.
func MultiPlexer(r *httprouter.Router) {
	Handle_CORE(r)
	INIT_AUTH_HANDLERS(r)
	INIT_OAUTH_Handlers(r)
	INIT_USERS_HANDLERS(r)
	INIT_NOTES_HANDLERS(r)
}

// Init do not touch.
func init() {
	r := httprouter.New()
	MultiPlexer(r)
	http.Handle("/", r)
}
