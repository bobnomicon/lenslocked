package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
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
		Index Template
		Show Template
	}

	Services struct {
		GalleryService *models.GalleryService
	}
}

/******** Helpers ********/

type galleryOpt func(w http.ResponseWriter, r *http.Request, gallery *models.Gallery) error

// Checks if gallery is owned by user, if not, writes a 404 Not Found status code and appropriate error message for the response.
func userMustOwnGallery(w http.ResponseWriter, r *http.Request, gallery *models.Gallery) error {
	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		w.WriteHeader(http.StatusNotFound)
		return apperrors.Public(ErrUnauthorized, "Gallery not found.")
	}

	return nil
}

// Checks if gallery is published, and if not owned by user, writes a 404 Not Found status code and appropriate error message for the response.
func galleryMustBePublished(w http.ResponseWriter, r *http.Request, gallery *models.Gallery) error {
	if !gallery.Published {
		user := context.User(r.Context())
		if user == nil || user.ID != gallery.UserID {
			w.WriteHeader(http.StatusNotFound)
			return apperrors.Public(ErrUnauthorized, "Gallery not found.")
		}
	}

	return nil
}

// Gets a gallery by ID from the request `id` param. If an error occurs, writes the appropriate status code and error message for the response. Additionally, if functional gallery options are passed, calls each one and returns an error if something is not okay.
func (g Galleries) galleryByID(w http.ResponseWriter, r *http.Request, opts ...galleryOpt) (*models.Gallery, error) {
	// Get the gallery id
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil, err
	}

	// Get the gallery
	gallery, err := g.Services.GalleryService.GetById(id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			err = apperrors.Public(err, "Gallery not found.")
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		return nil, err
	}

	// Iterate over functional gallery options, call each one and return error if there is an error
	for _, opt := range opts {
		if err = opt(w, r, gallery); err != nil {
			return nil, err
		}
	}

	return gallery, nil
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
	type Image struct {
		GalleryID int
		Filename string
		FilenameEscaped string
	}
	var data struct {
		ID int
		Title string
		Published bool
		Images []Image
	}

	// Get the gallery, check it belongs to user
	gallery, err := g.galleryByID(w, r, userMustOwnGallery)
	if err != nil {
		g.Templates.Edit.Execute(w, r, data, err)
		return
	}

	data.ID = gallery.ID
	data.Title = gallery.Title
	data.Published = gallery.Published

	images, err := g.Services.GalleryService.Images(gallery.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		g.Templates.Edit.Execute(w, r, data, err)
		return
	}

	for _, image := range images {
		data.Images = append(data.Images, Image{
			GalleryID: image.GalleryID,
			Filename: image.Filename,
			FilenameEscaped: url.PathEscape(image.Filename),
		})
	}

	g.Templates.Edit.Execute(w, r, data)
}

func (g Galleries) Index(w http.ResponseWriter, r *http.Request) {
	type Gallery struct {
		ID int
		Title string
		Published bool
	}
	var data struct {
		Galleries []Gallery
	}

	// Get all galleries that belong to user
	user := context.User(r.Context())
	galleries, err := g.Services.GalleryService.GetAllByUserID(user.ID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			// User has no galleries
			g.Templates.Index.Execute(w, r, data)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			g.Templates.Index.Execute(w, r, data, err)
		}

		return
	}

	for _, gallery := range galleries {
		data.Galleries = append(data.Galleries, Gallery{
			ID: gallery.ID,
			Title: gallery.Title,
			Published: gallery.Published,
		})
	}

	g.Templates.Index.Execute(w, r, data)
}

func (g Galleries) Show(w http.ResponseWriter, r *http.Request) {
	type Image struct {
		GalleryID int
		Filename string
		FilenameEscaped string
	}
	var data struct {
		ID int
		Title string
		Published bool
		Images []Image
	}

	// Get the gallery, check it is published or belongs to user
	gallery, err := g.galleryByID(w, r, galleryMustBePublished)
	if err != nil {
		g.Templates.Show.Execute(w, r, data, err)
		return
	}
	data.ID = gallery.ID
	data.Title = gallery.Title
	data.Published = gallery.Published

	images, err := g.Services.GalleryService.Images(gallery.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		g.Templates.Show.Execute(w, r, data, err)
		return
	}
	
	for _, image := range images {
		data.Images = append(data.Images, Image{
			GalleryID: image.GalleryID,
			Filename: image.Filename,
			FilenameEscaped: url.PathEscape(image.Filename),
		})
	}

	g.Templates.Show.Execute(w, r, data)
}

func (g Galleries) Image(w http.ResponseWriter, r *http.Request) {
	type Image struct {
		GalleryID int
		Filename string
	}
	var data struct {
		ID int
		Title string
		Published bool
		Images []Image
	}

	// Get the gallery, check it is published or belongs to user
	gallery, err := g.galleryByID(w, r, galleryMustBePublished)
	if err != nil {
		g.Templates.Show.Execute(w, r, data, err)
		return
	}
	data.ID = gallery.ID
	data.Title = gallery.Title
	data.Published = gallery.Published

	// Get the image
	filename := chi.URLParam(r, "filename")
	image, err := g.Services.GalleryService.Image(gallery.ID, filename)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			// Image does not exist
			w.WriteHeader(http.StatusNotFound)
			err = apperrors.Public(err, "Image not found.")
		} else {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		g.Templates.Show.Execute(w, r, data, err)
		return
	}

	http.ServeFile(w, r, image.Path)
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
		Published bool
	}

	// Get the gallery, check it belongs to user
	gallery, err := g.galleryByID(w, r, userMustOwnGallery)
	if err != nil {
		g.Templates.Edit.Execute(w, r, data, err)
		return
	}
	data.ID = gallery.ID

	// Parse form fields
	err = r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		g.Templates.Edit.Execute(w, r, data, err)
		return
	}
	data.Title = r.PostForm.Get("title")
	data.Published = r.PostForm.Get("published") == "on"

	// Check required fields
	if data.Title == "" {
		err = apperrors.Public(ErrMissingRequiredFields, "Please enter a title.")
		w.WriteHeader(http.StatusBadRequest)
		g.Templates.Edit.Execute(w, r, data, err)
		return
	}

	// Update the gallery
	gallery.Title = data.Title
	gallery.Published = data.Published
	err = g.Services.GalleryService.Update(gallery)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		g.Templates.Edit.Execute(w, r, data, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/galleries/%d/edit", gallery.ID), http.StatusFound)
}

func (g Galleries) Delete(w http.ResponseWriter, r *http.Request) {
	var data struct {
		ID int
		Title string
		Published bool
	}

	// Get the gallery, check it belongs to user
	gallery, err := g.galleryByID(w, r, userMustOwnGallery)
	if err != nil {
		g.Templates.Edit.Execute(w, r, data, err)
		return
	}
	data.ID = gallery.ID
	data.Title = gallery.Title
	data.Published = gallery.Published

	// Delete the gallery
	err = g.Services.GalleryService.Delete(gallery.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		g.Templates.Edit.Execute(w, r, data, err)
		return
	}

	http.Redirect(w, r, "/galleries", http.StatusFound)
}
