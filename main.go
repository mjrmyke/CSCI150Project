package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"text/template"
	"log"
)

func init() {
	r := httprouter.New()
	http.Handle("/", r)
	r.GET("/", func(res http.ResponseWriter, req *http.Request, params httprouter.Params){
		tpl, err := template.ParseFiles("index.gohtml")
		if err != nil {
			log.Fatalln(err)
		}

		err = tpl.Execute(res, nil)
		if err != nil {
			log.Fatalln(err)
		}

	})
}
