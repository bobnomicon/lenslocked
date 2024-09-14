package controllers

import (
	"fmt"
	"net/http"

	"github.com/operas-logicas/lenslocked/models"
)

type Users struct {
	Templates struct {
		SignUp Template
		SignIn Template
	}
	UserService *models.UserService
}

/******** GET handlers ********/

func (u Users) SignUp(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
	}

	data.Email = r.FormValue("email")
	u.Templates.SignUp.Execute(w, r, data)
}

func (u Users) SignIn(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
	}

	data.Email = r.FormValue("email")
	u.Templates.SignIn.Execute(w, r, data)
}

func (u Users) CurrentUser(w http.ResponseWriter, r *http.Request) {
	email, err := r.Cookie("email")
	if err != nil {
		fmt.Fprint(w, "No current user.")
		return
	}

	fmt.Fprintf(w, "Current user: %s\n", email.Value)
	fmt.Fprintf(w, "Headers: %+v\n", r.Header)
}


/******** POST handlers ********/

func (u Users) Create(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Please check required fields and try again.", http.StatusBadRequest)
		return
	}

	// Check confirm password and password match
	if r.PostForm.Get("password_confirm") != r.PostForm.Get("password") {
		http.Error(w, "Passwords do not match!", http.StatusBadRequest)
		return
	}

	user, err := u.UserService.Create(r.PostForm.Get("email"), r.PostForm.Get("password"))
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Please enter a valid email and password.", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "User created: %+v", user)
}

func (u Users) Authenticate(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, ".", http.StatusBadRequest)
	}

	user, err := u.UserService.Authenticate(r.PostForm.Get("email"), r.PostForm.Get("password"))
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Invalid credentials!", http.StatusUnauthorized)
	}

	cookie := http.Cookie{
		Name: "email",
		Value: user.Email,
		Path: "/",
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)

	fmt.Fprint(w, "User account authenticated!")
}
