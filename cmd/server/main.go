package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/bobnomicon/lenslocked/controllers"
	"github.com/bobnomicon/lenslocked/email"
	"github.com/bobnomicon/lenslocked/migrations"
	"github.com/bobnomicon/lenslocked/models"
	"github.com/bobnomicon/lenslocked/templates"
	"github.com/bobnomicon/lenslocked/views"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

type config struct {
	PSQL models.PostgresConfig
	SMTP email.SMTPConfig
	Server struct {
		Protocol string
		Address string
		Port string
	}
	App struct {
		Url string
		ImagesDir string
		ImagesExt string
		ImagesType string
	}
}

type server struct {
	Controllers struct {
		Static *controllers.Static
		Users *controllers.Users
		Galleries *controllers.Galleries
	}
	Middleware struct {
		Users *controllers.UserMiddleware
	}
}

func loadEnvConfig() (config, error) {
	var cfg config

	// Load .env
	if err := godotenv.Load(); err != nil {
		return cfg, err
	}

	// Set config
	cfg.PSQL = models.DefaultPostgresConfig()
	cfg.SMTP = email.DefaultSMTPConfig()
	cfg.Server.Address = os.Getenv("SERVER_ADDRESS")
	cfg.Server.Port = os.Getenv("SERVER_PORT")
	cfg.App.Url = os.Getenv("APP_URL")
	cfg.App.ImagesDir = os.Getenv("APP_IMAGES_DIR")
	cfg.App.ImagesExt = os.Getenv("APP_IMAGES_EXT")
	cfg.App.ImagesType = os.Getenv("APP_IMAGES_TYPE")

	return cfg, nil
}

func parseTemplates(s *server) {
	// Parse static templates
	s.Controllers.Static.Templates.Home = views.Must(views.ParseFS(templates.FS,
		"home.gohtml", "layout.gohtml",
	))
	s.Controllers.Static.Templates.Contact = views.Must(views.ParseFS(templates.FS,
		"contact.gohtml", "layout.gohtml",
	))
	s.Controllers.Static.Templates.FAQ = views.Must(views.ParseFS(templates.FS,
		"faq.gohtml", "layout.gohtml",
	))

	// Parse users templates
	s.Controllers.Users.Templates.SignUp = views.Must(views.ParseFS(templates.FS,
		"signup.gohtml", "layout.gohtml",
	))
	s.Controllers.Users.Templates.SignIn = views.Must(views.ParseFS(templates.FS,
		"signin.gohtml", "layout.gohtml",
	))
	s.Controllers.Users.Templates.ForgotPassword = views.Must(views.ParseFS(templates.FS,
		"forgot-password.gohtml", "layout.gohtml",
	))
	s.Controllers.Users.Templates.CheckEmail = views.Must(views.ParseFS(templates.FS,
		"check-email.gohtml", "layout.gohtml",
	))
	s.Controllers.Users.Templates.ResetPassword = views.Must(views.ParseFS(templates.FS,
		"reset-password.gohtml", "layout.gohtml",
	))
	s.Controllers.Users.Templates.ForgotPasswordEmail = email.Must(email.ParseFS(templates.FS,
		"emails/forgot-password.gohtml",
	))

	// Parse galleries templates
	s.Controllers.Galleries.Templates.New = views.Must(views.ParseFS(templates.FS,
		"galleries/new.gohtml", "layout.gohtml",
	))
	s.Controllers.Galleries.Templates.Edit = views.Must(views.ParseFS(templates.FS,
		"galleries/edit.gohtml", "layout.gohtml",
	))
	s.Controllers.Galleries.Templates.Index = views.Must(views.ParseFS(templates.FS,
		"galleries/index.gohtml", "layout.gohtml",
	))
	s.Controllers.Galleries.Templates.Show = views.Must(views.ParseFS(templates.FS,
		"galleries/show.gohtml", "layout.gohtml",
	))
}

func routes(r *chi.Mux, s *server) {
	// Static assets
	r.Get("/assets/*", controllers.AssetsHandler(http.Dir("assets")))

	// Static routes
	r.Get("/", controllers.StaticHandler(s.Controllers.Static.Templates.Home))
	r.Get("/contact", controllers.StaticHandler(s.Controllers.Static.Templates.Contact))	
	r.Get("/faq", controllers.FAQ(s.Controllers.Static.Templates.FAQ))

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

func run(cfg config) error {
	// Open DB connection	
	db, err := models.Open(cfg.PSQL)
	if err != nil {
		return err
	}
	defer db.Close()

	// Verify DB connection
	if err = db.Ping(); err != nil {
		return err
	}
	fmt.Println("Database connected")

	// Run migrations
	if err = models.MigrateFS(db, migrations.FS, "."); err != nil {
		return err
	}

	// Init services
	userService := &models.UserService{DB: db}
	sessionService := &models.SessionService{DB: db}
	passwordResetService := &models.PasswordResetService{DB: db}
	galleryService := &models.GalleryService{
		DB: db,
		ImagesDir: cfg.App.ImagesDir,
		ImagesExt: strings.Split(cfg.App.ImagesExt, ","),
		ImagesType: strings.Split(cfg.App.ImagesType, ","),
	}
	emailService := email.NewEmailService(cfg.SMTP)

	// Init controllers
	var staticController controllers.Static
	var usersController controllers.Users
	var galleriesController controllers.Galleries
	usersController.Services.UserService = userService
	usersController.Services.SessionService = sessionService
	usersController.Services.PasswordResetService = passwordResetService
	usersController.Services.EmailService = emailService
	galleriesController.Services.GalleryService = galleryService

	// Init middleware
	userMiddleware := controllers.UserMiddleware{
		SessionService: sessionService,
	}

	// Init server
	var server server
	server.Controllers.Static = &staticController
	server.Controllers.Users = &usersController
	server.Controllers.Galleries = &galleriesController
	server.Middleware.Users = &userMiddleware

	// Parse templates
	parseTemplates(&server)
	fmt.Println("Done parsing templates")

	// Init router
	r := chi.NewRouter()

	// CSRF Protection
	csrfProtect := http.NewCrossOriginProtection()
	csrfProtect.AddTrustedOrigin(cfg.Server.Address)
	
	// Global Middleware
	r.Use(
		middleware.Logger,
		csrfProtect.Handler,
		server.Middleware.Users.SetUser,
	)

	// Init routes
	routes(r, &server)

	// Start HTTP server and listen on specified port
	fmt.Printf("Starting the server on port %s...\n", cfg.Server.Port)
	addr := cfg.Server.Address + ":" + cfg.Server.Port
	return http.ListenAndServe(addr, r)
}

func main() {
	// Load env config
	cfg, err := loadEnvConfig()
	if err != nil {
		panic(err)
	}

	// Run the server
	if err = run(cfg); err != nil {
		panic(err)
	}
}
