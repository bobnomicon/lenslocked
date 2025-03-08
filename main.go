package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/csrf"
	"github.com/joho/godotenv"
	"github.com/operas-logicas/lenslocked/controllers"
	"github.com/operas-logicas/lenslocked/migrations"
	"github.com/operas-logicas/lenslocked/models"
	"github.com/operas-logicas/lenslocked/templates"
	"github.com/operas-logicas/lenslocked/views"
)

// Get the bool value of the CSRF_SECURE environment variable, default to true if error
func csrf_secure() bool {
	csrf_secure, err := strconv.ParseBool(os.Getenv("CSRF_SECURE"))
	if err != nil {
		return true
	}
	return csrf_secure
}

func main() {
	// Load .env
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	
	// Open DB connection
	cfg := models.DefaultPostgresConfig()
	// fmt.Println(cfg.String())
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

	// Run migrations
	err = models.MigrateFS(db, migrations.FS, ".")
	if err != nil {
		panic(err)
	}

	// Init services
	userService := models.UserService{DB: db}
	sessionService := models.SessionService{DB: db}

	userMiddleware := controllers.UserMiddleware{
		SessionService: &sessionService,
	}

	usersController := controllers.Users{
		UserService: &userService,
		SessionService: &sessionService,
	}

	// Parse static templates
	homeTemplate := views.Must(views.ParseFS(templates.FS, "home.gohtml", "layout.gohtml"))
	contactTemplate := views.Must(views.ParseFS(templates.FS, "contact.gohtml", "layout.gohtml"))
	faqTemplate := views.Must(views.ParseFS(templates.FS, "faq.gohtml", "layout.gohtml"))

	// Parse users templates
	usersController.Templates.SignUp = views.Must(views.ParseFS(templates.FS, "signup.gohtml", "layout.gohtml"))
	usersController.Templates.SignIn = views.Must(views.ParseFS(templates.FS, "signin.gohtml", "layout.gohtml"))

	// Init router
	r := chi.NewRouter()

	// Global Middlewares
	r.Use(
		middleware.Logger,
		csrf.Protect(
			[]byte(os.Getenv("CSRF_AUTH_KEY")),
			csrf.Secure(csrf_secure()),
		),
		userMiddleware.SetUser,
	)

	// Static routes
	r.Get("/", controllers.StaticHandler(homeTemplate))
	r.Get("/contact", controllers.StaticHandler(contactTemplate))	
	r.Get("/faq", controllers.FAQ(faqTemplate))

	// Users routes
	r.Get("/signin", usersController.SignIn)
	r.Post("/signin", usersController.Authenticate)
	r.Post("/signout", usersController.SignOut)
	r.Get("/signup", usersController.SignUp)
	r.Post("/signup", usersController.Create)
	
	r.Route("/users/me", func(r chi.Router) {
		r.Use(userMiddleware.RequireUser)
		r.Get("/", usersController.CurrentUser)
	})

	// 404 route
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})

	// Start HTTP server and listen on port 3000
	fmt.Println("Starting the server on :3000...")
	http.ListenAndServe("localhost:3000", r)
}
