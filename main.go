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
	homeTemplate := views.Must(views.ParseFS(templates.FS, "home.gohtml", "tailwind.gohtml"))
	contactTemplate := views.Must(views.ParseFS(templates.FS, "contact.gohtml", "tailwind.gohtml"))
	faqTemplate := views.Must(views.ParseFS(templates.FS, "faq.gohtml", "tailwind.gohtml"))
	signupTemplate := views.Must(views.ParseFS(templates.FS, "signup.gohtml", "tailwind.gohtml"))

	// Init router
	r := chi.NewRouter()	

	// Middlewares
	r.Use(middleware.Logger)

	// Routes
	r.Get("/", controllers.StaticHandler(homeTemplate))
	r.Get("/contact", controllers.StaticHandler(contactTemplate))	
	r.Get("/faq", controllers.FAQ(faqTemplate))
	r.Get("/signup", controllers.StaticHandler(signupTemplate))

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})

	// Start HTTP server and listen on port 3000
	fmt.Println("Starting the server on :3000...")
	http.ListenAndServe("localhost:3000", r)
}
