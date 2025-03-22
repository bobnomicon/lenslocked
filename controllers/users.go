package controllers

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/operas-logicas/lenslocked/context"
	"github.com/operas-logicas/lenslocked/email"
	"github.com/operas-logicas/lenslocked/models"
)

type Users struct {
	Templates struct {
		SignUp Template
		SignIn Template
		ForgotPassword Template
		CheckEmail Template
		ResetPassword Template
		ForgotPasswordEmail email.Template
	}

	Services struct {
		UserService *models.UserService
		SessionService *models.SessionService
		PasswordResetService *models.PasswordResetService
		EmailService *email.EmailService
	}
}

type UserMiddleware struct {
	SessionService *models.SessionService
}

/******** Middlewares ********/

func (umw UserMiddleware) SetUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := readCookie(r, CookieSession)
		if err != nil || token == "" {
			// Session cookie not set or token empty. Proceed with the request without setting user in the context.
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Error, empty session token.")
			}
			next.ServeHTTP(w, r)
			return
		}

		user, err := umw.SessionService.User(token)
		if err != nil {
			// Invalid or expired session token. Proceed with the request without setting user in the context.
			fmt.Println(err)
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

func (umw UserMiddleware) RequireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := context.User(r.Context())
		if user == nil {
			http.Redirect(w, r, "/signin", http.StatusFound)
			return
		}

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

// SetUser, RequireUser middleware required!
func (u Users) CurrentUser(w http.ResponseWriter, r *http.Request) {
	user := context.User(r.Context())
	fmt.Fprintf(w, "Current user: %s\n", user.Email)
}

func (u Users) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
	}

	data.Email = r.FormValue("email")
	u.Templates.ForgotPassword.Execute(w, r, data)
}

func (u Users) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Token string
	}

	data.Token = r.FormValue("token")
	u.Templates.ResetPassword.Execute(w, r, data)
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
	user, err := u.Services.UserService.Create(r.PostForm.Get("email"), r.PostForm.Get("password"))
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Please enter a valid email and password.", http.StatusInternalServerError)
		return
	}

	// User successfully created, so create new session for user
	session, err := u.Services.SessionService.Create(user.ID)
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
		http.Error(w, "Please check required fields and try again.", http.StatusBadRequest)
	}

	// Authenticate user
	user, err := u.Services.UserService.Authenticate(r.PostForm.Get("email"), r.PostForm.Get("password"))
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Invalid credentials!", http.StatusUnauthorized)
		return
	}

	// User successfully authenticated, so create new session for user
	session, err := u.Services.SessionService.Create(user.ID)
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
  err = u.Services.SessionService.Delete(token)
  if err != nil {
    fmt.Println(err)
    http.Error(w, "Something went wrong!", http.StatusInternalServerError)
    return
  }

  // Delete session cookie
  deleteCookie(w, CookieSession)
  http.Redirect(w, r, "/signin", http.StatusFound)
}

func (u Users) ProcessForgotPassword(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Please enter a valid email.", http.StatusBadRequest)
		return
	}

	var data struct {
		Email string
	}
	data.Email = r.PostForm.Get("email")

	passwordReset, err := u.Services.PasswordResetService.Create(data.Email)
	if err != nil {
		// TODO: Handle case where a user with that email address doesn't exist.
		fmt.Println(err)
		http.Error(w, "Something went wrong!", http.StatusInternalServerError)
		return
	}

	vals := url.Values{}
	vals.Set("token", passwordReset.Token)

	// TODO: Make URL configurable.
	resetURL := "https://www.lenslocked.com/reset-password?" + vals.Encode()
	err = u.Services.EmailService.ForgotPassword(data.Email, resetURL, &u.Templates.ForgotPasswordEmail)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong!", http.StatusInternalServerError)
		return
	}

	u.Templates.CheckEmail.Execute(w, r, data)
}

func (u Users) ProcessResetPassword(w http.ResponseWriter, r *http.Request) {
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

	var data struct {
		Token string
		Password string
	}
	data.Token = r.PostForm.Get("token")
	data.Password = r.PostForm.Get("password")

	user, err := u.Services.PasswordResetService.Consume(data.Token)
	if err != nil {
		// TODO: Handle invalid token errors.
		fmt.Println(err)
		http.Error(w, "Something went wrong!", http.StatusInternalServerError)
		return
	}

	// Update user's password
	err = u.Services.UserService.UpdatePassword(user.ID, data.Password)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong!", http.StatusInternalServerError)
		return
	}

	// Sign user in now that they have reset their password by creating a new session for user
	session, err := u.Services.SessionService.Create(user.ID)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}

	// Set cookie with session token
  setCookie(w, CookieSession, session.Token)
	http.Redirect(w, r, "/users/me", http.StatusFound)
}