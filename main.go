package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/operas-logicas/lenslocked/controllers"
	"github.com/operas-logicas/lenslocked/templates"
	"github.com/operas-logicas/lenslocked/views"
)

func main() {
	// Parse templates
	homeTemplate := views.Must(views.ParseFS(templates.FS, "layout-page.gohtml", "home-page.gohtml"))
	contactTemplate := views.Must(views.ParseFS(templates.FS, "layout-page.gohtml", "contact-page.gohtml"))
	faqTemplate := views.Must(views.ParseFS(templates.FS, "layout-page.gohtml", "faq-page.gohtml"))

	// Init router
	r := chi.NewRouter()	

	// Middlewares
	r.Use(middleware.Logger)

	// Routes
	r.Get("/", controllers.StaticHandler(homeTemplate))
	r.Get("/contact", controllers.StaticHandler(contactTemplate))	
	r.Get("/faq", controllers.FAQ(faqTemplate))

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})

	// Start HTTP server and listen on port 3000
	fmt.Println("Starting the server on :3000...")
	http.ListenAndServe("localhost:3000", r)
}
