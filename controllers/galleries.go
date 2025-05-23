package controllers

import (
	"fmt"
	"net/http"

	"github.com/bobnomicon/lenslocked/context"
	apperrors "github.com/bobnomicon/lenslocked/errors"
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


/******** POST handlers ********/

func (g Galleries) Create(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Title string
	}

	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		g.Templates.New.Execute(w, r, data, err)
		return
	}

	data.Title = r.PostForm.Get("title")

	// Check required fields
	if data.Title == "" {
		err = apperrors.Public(ErrMissingRequiredFields, "Please enter a title.")
		w.WriteHeader(http.StatusBadRequest)
		g.Templates.New.Execute(w, r, data, err)
		return
	}

	gallery, err := g.Services.GalleryService.Create(data.Title, context.User(r.Context()).ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		g.Templates.New.Execute(w, r, data, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/galleries/%d/edit", gallery.ID), http.StatusFound)
}
