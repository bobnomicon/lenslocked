package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/bobnomicon/lenslocked/controllers"
	"github.com/bobnomicon/lenslocked/email"
	"github.com/bobnomicon/lenslocked/migrations"
	"github.com/bobnomicon/lenslocked/models"
	"github.com/bobnomicon/lenslocked/templates"
	"github.com/bobnomicon/lenslocked/views"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/csrf"
	"github.com/joho/godotenv"
)

type config struct {
	PSQL models.PostgresConfig
	SMTP email.SMTPConfig
	CSRF struct {
		Key string
		Secure bool
	}
	Server struct {
		Protocol string
		Address string
		Port string
	}
	App struct {
		Url string
	}
}

func loadEnvConfig() (config, error) {
	var cfg config

	// Load .env
	if err := godotenv.Load(); err != nil {
		return cfg, err
	}

	// Get the bool value of the CSRF_SECURE environment variable, default to true if error
	csrfSecure, err := strconv.ParseBool(os.Getenv("CSRF_SECURE"))
	if err != nil {
		csrfSecure = true
	}

	// Set config
	cfg.PSQL = models.DefaultPostgresConfig()
	cfg.SMTP = email.DefaultSMTPConfig()
	cfg.CSRF.Key = os.Getenv("CSRF_AUTH_KEY")
	cfg.CSRF.Secure = csrfSecure
	cfg.Server.Address = os.Getenv("SERVER_ADDRESS")
	cfg.Server.Port = os.Getenv("SERVER_PORT")
	cfg.App.Url = os.Getenv("APP_URL")

	return cfg, nil
}

func main() {
	// Load env config
	cfg, err := loadEnvConfig()
	if err != nil {
		panic(err)
	}

	// Open DB connection	
	db, err := models.Open(cfg.PSQL)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Verify DB connection
	if err = db.Ping(); err != nil {
		panic(err)
	}
	fmt.Println("Database connected")

	// Run migrations
	if err = models.MigrateFS(db, migrations.FS, "."); err != nil {
		panic(err)
	}

	// Init services
	userService := &models.UserService{DB: db}
	sessionService := &models.SessionService{DB: db}
	passwordResetService := &models.PasswordResetService{DB: db}
	galleryService := &models.GalleryService{DB: db}
	emailService := email.NewEmailService(cfg.SMTP)

	var usersController controllers.Users
	usersController.Services.UserService = userService
	usersController.Services.SessionService = sessionService
	usersController.Services.PasswordResetService = passwordResetService
	usersController.Services.EmailService = emailService

	userMiddleware := controllers.UserMiddleware{
		SessionService: sessionService,
	}

	var galleriesController controllers.Galleries
	galleriesController.Services.GalleryService = galleryService

	// Parse static templates
	homeTemplate := views.Must(views.ParseFS(templates.FS, "home.gohtml", "layout.gohtml"))
	contactTemplate := views.Must(views.ParseFS(templates.FS, "contact.gohtml", "layout.gohtml"))
	faqTemplate := views.Must(views.ParseFS(templates.FS, "faq.gohtml", "layout.gohtml"))

	// Parse users templates
	usersController.Templates.SignUp = views.Must(views.ParseFS(templates.FS, "signup.gohtml", "layout.gohtml"))
	usersController.Templates.SignIn = views.Must(views.ParseFS(templates.FS, "signin.gohtml", "layout.gohtml"))
	usersController.Templates.ForgotPassword = views.Must(views.ParseFS(templates.FS, "forgot-password.gohtml", "layout.gohtml"))
	usersController.Templates.CheckEmail = views.Must(views.ParseFS(templates.FS, "check-email.gohtml", "layout.gohtml"))
	usersController.Templates.ResetPassword = views.Must(views.ParseFS(templates.FS, "reset-password.gohtml", "layout.gohtml"))
	usersController.Templates.ForgotPasswordEmail = email.Must(email.ParseFS(templates.FS, "emails/forgot-password.gohtml"))

	// Parse galleries templates
	galleriesController.Templates.New = views.Must(views.ParseFS(templates.FS, "galleries/new.gohtml", "layout.gohtml"))
	galleriesController.Templates.Edit = views.Must(views.ParseFS(templates.FS, "galleries/edit.gohtml", "layout.gohtml"))

	fmt.Println("Done parsing templates")

	// Init router
	r := chi.NewRouter()

	// Global Middlewares
	r.Use(
		middleware.Logger,
		csrf.Protect(
			[]byte(cfg.CSRF.Key),
			csrf.Secure(cfg.CSRF.Secure),
			csrf.Path("/"),
		),
		userMiddleware.SetUser,
	)

	// Static routes
	r.Get("/", controllers.StaticHandler(homeTemplate))
	r.Get("/contact", controllers.StaticHandler(contactTemplate))	
	r.Get("/faq", controllers.FAQ(faqTemplate))

	// Users routes
	r.Get("/forgot-password", usersController.ForgotPassword)
	r.Post("/forgot-password", usersController.ProcessForgotPassword)
	r.Get("/reset-password", usersController.ResetPassword)
	r.Post("/reset-password", usersController.ProcessResetPassword)
	r.Get("/signin", usersController.SignIn)
	r.Post("/signin", usersController.Authenticate)
	r.Post("/signout", usersController.SignOut)
	r.Get("/signup", usersController.SignUp)
	r.Post("/signup", usersController.Create)

	// Users routes - REQUIRE USER
	r.Route("/users/me", func(r chi.Router) {
		r.Use(userMiddleware.RequireUser)
		r.Get("/", usersController.CurrentUser)
	})

	// Galleries routes
	r.Route("/galleries", func(r chi.Router) {
		// REQUIRE USER
		r.Group(func(r chi.Router) {
			r.Use(userMiddleware.RequireUser)
			r.Get("/new", galleriesController.New)
			r.Post("/", galleriesController.Create)
			r.Get("/{id}/edit", galleriesController.Edit)
		})
	})

	// 404 route
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})

	// Start HTTP server and listen on specified port
	fmt.Printf("Starting the server on port %s...\n", cfg.Server.Port)
	addr := cfg.Server.Address + ":" + cfg.Server.Port
	if err = http.ListenAndServe(addr, r); err != nil {
		panic(err)
	}
}
