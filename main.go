package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"fmt"
)

func init() {
	r := httprouter.New()
	http.Handle("/", r)
	r.GET("/", func(res http.ResponseWriter, req *http.Request, params httprouter.Params){
		fmt.Fprint(res,"Hello World")
	})
}
