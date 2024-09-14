package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/csrf"
	"github.com/joho/godotenv"
	"github.com/operas-logicas/lenslocked/controllers"
	"github.com/operas-logicas/lenslocked/models"
	"github.com/operas-logicas/lenslocked/templates"
	"github.com/operas-logicas/lenslocked/views"
)

func main() {
	// Load .env
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	
	// Open DB connection
	cfg := models.DefaultPostgresConfig()
	db, err := models.Open(cfg)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Verify DB connection
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("Database connected")

	// Init model services
	usersController := controllers.Users{
		UserService: &models.UserService{
			DB: db,
		},
	}

	// Parse static templates
	homeTemplate := views.Must(views.ParseFS(templates.FS, "home.gohtml", "tailwind.gohtml"))
	contactTemplate := views.Must(views.ParseFS(templates.FS, "contact.gohtml", "tailwind.gohtml"))
	faqTemplate := views.Must(views.ParseFS(templates.FS, "faq.gohtml", "tailwind.gohtml"))

	// Parse users templates
	usersController.Templates.SignUp = views.Must(views.ParseFS(templates.FS, "signup.gohtml", "tailwind.gohtml"))
	usersController.Templates.SignIn = views.Must(views.ParseFS(templates.FS, "signin.gohtml", "tailwind.gohtml"))

	// Init router
	r := chi.NewRouter()	

	// Middlewares
	csrfMiddleware := csrf.Protect(
		[]byte(os.Getenv("CSRF_AUTH_KEY")),
		// TODO! Fix before deploying to production:
		csrf.Secure(false),
	)
	r.Use(csrfMiddleware)
	r.Use(middleware.Logger)

	// Static routes
	r.Get("/", controllers.StaticHandler(homeTemplate))
	r.Get("/contact", controllers.StaticHandler(contactTemplate))	
	r.Get("/faq", controllers.FAQ(faqTemplate))

	// Users routes
	r.Get("/signin", usersController.SignIn)
	r.Post("/signin", usersController.Authenticate)
	r.Get("/signup", usersController.SignUp)
	r.Post("/signup", usersController.Create)
	r.Get("/users/me", usersController.CurrentUser)

	// 404 route
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})

	// Start HTTP server and listen on port 3000
	fmt.Println("Starting the server on :3000...")
	http.ListenAndServe("localhost:3000", r)
}
