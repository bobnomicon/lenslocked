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
	"github.com/operas-logicas/lenslocked/email"
	"github.com/operas-logicas/lenslocked/migrations"
	"github.com/operas-logicas/lenslocked/models"
	"github.com/operas-logicas/lenslocked/templates"
	"github.com/operas-logicas/lenslocked/views"
)

type config struct {
	PSQL models.PostgresConfig
	SMTP email.SMTPConfig
	CSRF struct {
		Key string
		Secure bool
	}
	Server struct {
		Address string
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
	emailService := email.NewEmailService(cfg.SMTP)

	var usersController controllers.Users
	usersController.Services.UserService = userService
	usersController.Services.SessionService = sessionService
	usersController.Services.PasswordResetService = passwordResetService
	usersController.Services.EmailService = emailService

	userMiddleware := controllers.UserMiddleware{
		SessionService: sessionService,
	}

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

	fmt.Println("Done parsing templates")

	// Init router
	r := chi.NewRouter()

	// Global Middlewares
	r.Use(
		middleware.Logger,
		csrf.Protect(
			[]byte(cfg.CSRF.Key),
			csrf.Secure(cfg.CSRF.Secure),
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
	
	r.Route("/users/me", func(r chi.Router) {
		r.Use(userMiddleware.RequireUser)
		r.Get("/", usersController.CurrentUser)
	})

	// 404 route
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})

	// Start HTTP server and listen on port 3000
	fmt.Printf("Starting the server on %s...\n", cfg.Server.Address)
	if err = http.ListenAndServe(cfg.Server.Address, r); err != nil {
		panic(err)
	}
}
