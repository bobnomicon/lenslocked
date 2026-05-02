package controllers

import (
	"fmt"
	"net/http"
)

type Static struct {
	Templates struct {
		Home Template
	}
}

func AssetsHandler(fs http.FileSystem) http.HandlerFunc {
	return http.StripPrefix(
		fmt.Sprintf("/%s", fs),
		http.FileServer(fs),
	).ServeHTTP
}

func StaticHandler(tpl Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tpl.Execute(w, r, nil)
	}
}
