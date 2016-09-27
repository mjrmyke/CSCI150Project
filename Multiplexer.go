package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

// Reference Multiplexer functions in other modules here.
// Those functions splinter off the handlers.
func MultiPlexer(r *httprouter.Router){
	Handle_CORE(r)
}

// Init do not touch.
func init() {
	r := httprouter.New()
	MultiPlexer(r)
	http.Handle("/", r)
}
