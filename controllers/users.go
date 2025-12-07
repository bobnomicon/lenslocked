package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/bobnomicon/lenslocked/context"
	"github.com/bobnomicon/lenslocked/email"
	apperrors "github.com/bobnomicon/lenslocked/errors"
	"github.com/bobnomicon/lenslocked/models"
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


/******** Middleware ********/

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
	var data struct {
		Email string
		Password string
		ConfirmPassword string
	}

	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		u.Templates.SignUp.Execute(w, r, data, err)
		return
	}

	data.Email = r.PostForm.Get("email")
	data.Password = r.PostForm.Get("password")
	data.ConfirmPassword = r.PostForm.Get("confirm_password")

	// Check required fields
	if data.Email == "" || data.Password == "" || data.ConfirmPassword == "" {
		err = apperrors.Public(ErrMissingRequiredFields, "Please check required fields and try again.")
		w.WriteHeader(http.StatusBadRequest)
		u.Templates.SignUp.Execute(w, r, data, err)
		return
	}

	// Check confirm password and password match
	if data.ConfirmPassword != data.Password {
		w.WriteHeader(http.StatusBadRequest)
		err = apperrors.Public(ErrPasswordsDontMatch, "Password and confirm password do not match.")
		u.Templates.SignUp.Execute(w, r, data, err)
		return
	}

	// Create user
	user, err := u.Services.UserService.Create(data.Email, data.Password)
	if err != nil {
		if errors.Is(err, models.ErrEmailTaken) {
			w.WriteHeader(http.StatusBadRequest)
			err = apperrors.Public(err, "Email address is already associated with an account.")
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		u.Templates.SignUp.Execute(w, r, data, err)
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
	http.Redirect(w, r, "/galleries", http.StatusFound)
}

func (u Users) Authenticate(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
		Password string
	}

	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		u.Templates.SignIn.Execute(w, r, data, err)
	}

	data.Email = r.PostForm.Get("email")
	data.Password = r.PostForm.Get("password")

	// Check required fields
	if data.Email == "" || data.Password == "" {
		err = apperrors.Public(ErrMissingRequiredFields, "Please check required fields and try again.")
		w.WriteHeader(http.StatusBadRequest)
		u.Templates.SignIn.Execute(w, r, data, err)
		return
	}

	// Authenticate user
	user, err := u.Services.UserService.Authenticate(data.Email, data.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			w.WriteHeader(http.StatusUnauthorized)
			err = apperrors.Public(err, "Please enter a valid email and password.")
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		
		u.Templates.SignIn.Execute(w, r, data, err)
		return
	}

	// User successfully authenticated, so create new session for user
	session, err := u.Services.SessionService.Create(user.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		u.Templates.SignIn.Execute(w, r, data, err)
		return
	}

	// Set cookie with session token
  setCookie(w, CookieSession, session.Token)
	http.Redirect(w, r, "/galleries", http.StatusFound)
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
		// Don't return (delete the cookie and redirect to sign in page)
  }

  // Delete session cookie
  deleteCookie(w, CookieSession)
  http.Redirect(w, r, "/signin", http.StatusFound)
}

func (u Users) ProcessForgotPassword(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
	}

	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		u.Templates.ForgotPassword.Execute(w, r, data, err)
		return
	}

	data.Email = r.PostForm.Get("email")

	// Check required fields
	if data.Email == "" {
		err = apperrors.Public(ErrMissingRequiredFields, "Please a enter a valid email.")
		w.WriteHeader(http.StatusBadRequest)
		u.Templates.ForgotPassword.Execute(w, r, data, err)
		return
	}

	passwordReset, err := u.Services.PasswordResetService.Create(data.Email)
	if err != nil {
		if errors.Is(err, models.ErrEmailDoesNotExist) {
			w.WriteHeader(http.StatusBadRequest)
			err = apperrors.Public(err, "An account with that email address does not exist.")
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		
		u.Templates.ForgotPassword.Execute(w, r, data, err)
		return
	}

	vals := url.Values{}
	vals.Set("token", passwordReset.Token)

	resetURL := os.Getenv("APP_URL") +  "/reset-password?" + vals.Encode()
	err = u.Services.EmailService.ForgotPassword(data.Email, resetURL, &u.Templates.ForgotPasswordEmail)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		u.Templates.ForgotPassword.Execute(w, r, data, err)
		return
	}

	u.Templates.CheckEmail.Execute(w, r, data)
}

func (u Users) ProcessResetPassword(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Token string
		Password string
		ConfirmPassword string
	}

	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		u.Templates.ResetPassword.Execute(w, r, data, err)
		return
	}

	data.Token = r.PostForm.Get("token")
	data.Password = r.PostForm.Get("password")
	data.ConfirmPassword = r.PostForm.Get("confirm_password")

	// Check required fields
	if data.Password == "" || data.ConfirmPassword == "" {
		err = apperrors.Public(ErrMissingRequiredFields, "Please check required fields and try again.")
		w.WriteHeader(http.StatusBadRequest)
		u.Templates.ResetPassword.Execute(w, r, data, err)
		return
	}

	// Check confirm password and password match
	if data.ConfirmPassword != data.Password {
		w.WriteHeader(http.StatusBadRequest)
		err = apperrors.Public(ErrPasswordsDontMatch, "Password and confirm password do not match.")
		u.Templates.ResetPassword.Execute(w, r, data, err)
		return
	}

	user, err := u.Services.PasswordResetService.Consume(data.Token)
	if err != nil {
		if errors.Is(err, models.ErrTokenInvalid) {
			w.WriteHeader(http.StatusBadRequest)
			err = apperrors.Public(err, "Password reset token is invalid.")
		} else if errors.Is(err, models.ErrTokenExpired) {
			w.WriteHeader(http.StatusBadRequest)
			err = apperrors.Public(err, "Password reset token has expired.")
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		u.Templates.ResetPassword.Execute(w, r, data, err)
		return
	}

	// Update user's password
	err = u.Services.UserService.UpdatePassword(user.ID, data.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		u.Templates.ResetPassword.Execute(w, r, data, err)
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
	http.Redirect(w, r, "/galleries", http.StatusFound)
}
