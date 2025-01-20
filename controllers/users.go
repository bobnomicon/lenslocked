package controllers

import (
	"fmt"
	"net/http"

	"github.com/operas-logicas/lenslocked/context"
	"github.com/operas-logicas/lenslocked/models"
)

type Users struct {
	Templates struct {
		SignUp Template
		SignIn Template
	}
	UserService *models.UserService
	SessionService *models.SessionService
}

type UserMiddleware struct {
	SessionService *models.SessionService
}

/******** Middlewares ********/
func (umw UserMiddleware) SetUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := readCookie(r, CookieSession)
		if err != nil {
			// Session cookie not set. Proceed with the request without setting user in the context.
			next.ServeHTTP(w, r)
			return
		}

		user, err := umw.SessionService.User(token)
		if err != nil {
			// Invalid or expired session token. Proceed with the request without setting user in the context.
			next.ServeHTTP(w, r)
			return
		}

		// Store user in the context!
		ctx := r.Context()
		ctx = context.WithUser(ctx, user)
		r = r.WithContext(ctx)
		
		// Proceed with the request with user set in the context.
		next.ServeHTTP(w, r)
	})
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
	user := context.User(r.Context())
	if user == nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	fmt.Fprintf(w, "Current user: %s\n", user.Email)
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

	// Create user
	user, err := u.UserService.Create(r.PostForm.Get("email"), r.PostForm.Get("password"))
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Please enter a valid email and password.", http.StatusInternalServerError)
		return
	}

	// User successfully created, so create new session for user
	session, err := u.SessionService.Create(user.ID)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	// Set cookie with session token
  setCookie(w, CookieSession, session.Token)
	http.Redirect(w, r, "/users/me", http.StatusFound)
}

func (u Users) Authenticate(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, ".", http.StatusBadRequest)
	}

	// Authenticate user
	user, err := u.UserService.Authenticate(r.PostForm.Get("email"), r.PostForm.Get("password"))
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Invalid credentials!", http.StatusUnauthorized)
		return
	}

	// User successfully authenticated, so create new session for user
	session, err := u.SessionService.Create(user.ID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong!", http.StatusInternalServerError)
		return
	}

	// Set cookie with session token
  setCookie(w, CookieSession, session.Token)
	http.Redirect(w, r, "/users/me", http.StatusFound)
}

func (u Users) SignOut(w http.ResponseWriter, r *http.Request) {
  token, err := readCookie(r, CookieSession)
  if err != nil {
    fmt.Println(err)
    http.Redirect(w, r, "/signin", http.StatusFound)
    return
  }

  // Delete user's session
  err = u.SessionService.Delete(token)
  if err != nil {
    fmt.Println(err)
    http.Error(w, "Something went wrong!", http.StatusInternalServerError)
    return
  }

  // Delete session cookie
  deleteCookie(w, CookieSession)
  http.Redirect(w, r, "/signin", http.StatusFound)
}
