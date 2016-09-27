package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func Handle_CORE(r *httprouter.Router){
	r.GET("/",index)
}

func index(res http.ResponseWriter, req *http.Request, p httprouter.Params){
	ServeTemplateWithParams(res, "index", nil)
}