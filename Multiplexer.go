package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func MultiPlexer(r *httprouter.Router){
	Handle_CORE(r)
}

func init() {
	r := httprouter.New()
	MultiPlexer(r)
	http.Handle("/", r)
}
