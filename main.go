package main

import (
	"fmt"
	"net/http"
	"path/filepath"

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
	tpl, err := views.Template{}.Parse(filepath.Join("templates", "home.gohtml"))
	if err != nil {
		panic(err)
	}
	// r.Method(http.MethodGet, "/", controllers.Static{ Template: tpl })
	r.Get("/", controllers.StaticHandler(tpl))
	
	tpl, err = views.Template{}.Parse(filepath.Join("templates", "contact.gohtml"))
	if err != nil {
		panic(err)
	}
	r.Get("/contact", controllers.StaticHandler(tpl))

	tpl, err = views.Template{}.Parse(filepath.Join("templates", "faq.gohtml"))
	if err != nil {
		panic(err)
	}
	r.Get("/faq", controllers.StaticHandler(tpl))

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})

	// Start HTTP server and listen on port 3000
	fmt.Println("Starting the server on :3000...")
	http.ListenAndServe("localhost:3000", r)
}
