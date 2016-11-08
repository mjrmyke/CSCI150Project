package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Multiplexer Function for CORE
func Handle_CORE(r *httprouter.Router) {
	r.GET("/", index)
}

// Serves the index page.
func index(res http.ResponseWriter, req *http.Request, p httprouter.Params) {
	ServeTemplateWithParams(res, "index.gohtml", struct {
		HeaderData
	}{
		*MakeHeader(res, req, true, true),
	})
}
