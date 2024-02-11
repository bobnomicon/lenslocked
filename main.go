package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/operas-logicas/lenslocked/controllers"
	"github.com/operas-logicas/lenslocked/views"
)

func main() {
	r := chi.NewRouter()	

	// Middlewares
	r.Use(middleware.Logger)

	// Routes
	r.Get("/", controllers.StaticHandler(views.Must(views.Parse("templates/home.gohtml"))))
	r.Get("/contact", controllers.StaticHandler(views.Must(views.Parse("templates/contact.gohtml"))))	
	r.Get("/faq", controllers.StaticHandler(views.Must(views.Parse("templates/faq.gohtml"))))

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})

	// Start HTTP server and listen on port 3000
	fmt.Println("Starting the server on :3000...")
	http.ListenAndServe("localhost:3000", r)
}
