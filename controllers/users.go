package controllers

import (
	"fmt"
	"net/http"

	"github.com/operas-logicas/lenslocked/models"
)

type Users struct {
	Templates struct {
		New Template
	}
	UserService *models.UserService
}

func (u Users) New(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
	}

	data.Email = r.FormValue("email")
	u.Templates.New.Execute(w, data)
}

func (u Users) Create(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Unable to parse form for submission.", http.StatusBadRequest)
		return
	}

	if r.PostForm.Get("password_confirm") != r.PostForm.Get("password") {
		http.Error(w, "Passwords do not match!", http.StatusBadRequest)
		return
	}

	user, err := u.UserService.Create(r.PostForm.Get("email"), r.PostForm.Get("password"))
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong!", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "User created: %+v", user)
}
