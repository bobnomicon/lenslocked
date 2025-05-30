package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/bobnomicon/lenslocked/context"
	apperrors "github.com/bobnomicon/lenslocked/errors"
	"github.com/bobnomicon/lenslocked/models"
	"github.com/go-chi/chi/v5"
)

type Galleries struct {
	Templates struct {
		New Template
		Edit Template
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

func (g Galleries) Edit(w http.ResponseWriter, r *http.Request) {
	var data struct {
		ID int
		Title string
	}

	// Get the gallery id
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		g.Templates.Edit.Execute(w, r, data, err)
		return
	}
	data.ID = id

	// Get the gallery
	gallery, err := g.Services.GalleryService.GetById(data.ID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			err = apperrors.Public(err, "Gallery not found.")
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		g.Templates.Edit.Execute(w, r, data, err)
		return
	}

	// Check gallery belongs to user
	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		w.WriteHeader(http.StatusForbidden)
		err = apperrors.Public(err, "You do not have permission to edit this gallery.")
		g.Templates.Edit.Execute(w, r, data, err)
		return
	}

	data.ID = gallery.ID
	data.Title = gallery.Title
	g.Templates.Edit.Execute(w, r, data)
}


/******** POST handlers ********/

func (g Galleries) Create(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Title string
	}

	// Parse form fields
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

	// Create the gallery
	gallery, err := g.Services.GalleryService.Create(data.Title, context.User(r.Context()).ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		g.Templates.New.Execute(w, r, data, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/galleries/%d/edit", gallery.ID), http.StatusFound)
}

func (g Galleries) Update(w http.ResponseWriter, r *http.Request) {
	var data struct {
		ID int
		Title string
	}

	// Get the gallery id
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		g.Templates.Edit.Execute(w, r, data, err)
		return
	}
	data.ID = id

	// Get the gallery
	gallery, err := g.Services.GalleryService.GetById(id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			err = apperrors.Public(err, "Gallery not found.")
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		g.Templates.Edit.Execute(w, r, data, err)
		return
	}

	// Check gallery belongs to user
	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		w.WriteHeader(http.StatusForbidden)
		err = apperrors.Public(err, "You do not have permission to edit this gallery.")
		g.Templates.Edit.Execute(w, r, data, err)
		return
	}

	// Parse form fields
	err = r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		g.Templates.Edit.Execute(w, r, data, err)
		return
	}
	data.Title = r.PostForm.Get("title")

	// Check required fileds
	if data.Title == "" {
		err = apperrors.Public(ErrMissingRequiredFields, "Please enter a title.")
		w.WriteHeader(http.StatusBadRequest)
		g.Templates.Edit.Execute(w, r, data, err)
		return
	}

	// Update the gallery
	gallery.Title = data.Title
	err = g.Services.GalleryService.Update(gallery)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		g.Templates.Edit.Execute(w, r, data, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/galleries/%d/edit", gallery.ID), http.StatusFound)
}
