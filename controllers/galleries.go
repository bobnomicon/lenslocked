package controllers

import (
	"net/http"

	"github.com/bobnomicon/lenslocked/models"
)

type Galleries struct {
	Templates struct {
		New Template
	}

	Services struct {
		GalleryService *models.GalleryService
	}
}

/******** GET handlers ********/

func (g Galleries) New(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Title string
	}

	data.Title = r.FormValue("title")
	g.Templates.New.Execute(w, r, data)
}
