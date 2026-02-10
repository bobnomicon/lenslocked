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

	// Load .env (optional - ignore error if env vars already set)
	_ = godotenv.Load()

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

	// Start HTTP server and listen on specified port
	fmt.Printf("Starting server on port %s...\n", cfg.Server.Port)
	addr := cfg.Server.Address + ":" + cfg.Server.Port
	return http.ListenAndServe(addr, router(&server, cfg))
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
