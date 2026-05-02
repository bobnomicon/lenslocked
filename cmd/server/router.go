package main

import (
	"net/http"

	"github.com/bobnomicon/lenslocked/controllers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func routes(r *chi.Mux, s *server) {
	// Static assets
	r.Get("/assets/*", controllers.AssetsHandler(http.Dir("assets")))

	// Static routes
	r.Get("/", controllers.StaticHandler(s.Controllers.Static.Templates.Home))

	// Users routes
	r.Get("/forgot-password", s.Controllers.Users.ForgotPassword)
	r.Post("/forgot-password", s.Controllers.Users.ProcessForgotPassword)
	r.Get("/reset-password", s.Controllers.Users.ResetPassword)
	r.Post("/reset-password", s.Controllers.Users.ProcessResetPassword)
	r.Get("/signin", s.Controllers.Users.SignIn)
	r.Post("/signin", s.Controllers.Users.Authenticate)
	r.Post("/signout", s.Controllers.Users.SignOut)
	r.Get("/signup", s.Controllers.Users.SignUp)
	r.Post("/signup", s.Controllers.Users.Create)

	// Users routes - REQUIRE USER
	r.Route("/users/me", func(r chi.Router) {
		r.Use(s.Middleware.Users.RequireUser)
		r.Get("/", s.Controllers.Users.CurrentUser)
	})

	// Galleries routes
	r.Route("/galleries", func(r chi.Router) {
		// Anyone can view a gallery as long as it's published
		r.Get("/{id}", s.Controllers.Galleries.Show)
		r.Get("/{id}/images/{filename}", s.Controllers.Galleries.Image)

		// REQUIRE USER
		r.Group(func(r chi.Router) {
			r.Use(s.Middleware.Users.RequireUser)
			r.Get("/", s.Controllers.Galleries.Index)
			r.Get("/new", s.Controllers.Galleries.New)
			r.Post("/new", s.Controllers.Galleries.Create)
			r.Get("/{id}/edit", s.Controllers.Galleries.Edit)
			r.Post("/{id}/edit", s.Controllers.Galleries.Update)
			r.Post("/{id}/delete", s.Controllers.Galleries.Delete)
			r.Post("/{id}/images", s.Controllers.Galleries.UploadImage)
			r.Post("/{id}/images/{filename}/delete", s.Controllers.Galleries.DeleteImage)
		})
	})

	// 404 route
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})
}

// Initializes the router, global middleware, and routes
func router(s *server, cfg config) http.Handler {
	// Init router
	r := chi.NewRouter()

	// CSRF Protection
	csrfProtect := http.NewCrossOriginProtection()
	csrfProtect.AddTrustedOrigin(cfg.App.Url)

	// Global Middleware
	r.Use(
		middleware.Logger,
		csrfProtect.Handler,
		s.Middleware.Users.SetUser,
	)

	// Init routes
	routes(r, s)

	return r
}
