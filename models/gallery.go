package models

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Gallery struct {
	ID int
	UserID int
	Title string
	Published bool
}

type Image struct {
	GalleryID int
	Path string
	Filename string
}

type GalleryService struct {
	DB *sql.DB
	ImagesDir string // Directory to store and locate images
	ImagesExt []string // Supported image file extensions
	ImagesType []string // Supported image file content types
}

func hasExtension(file string, extensions ...string) bool {
	file = strings.ToLower(file)

	for _, ext := range extensions {
		ext = strings.ToLower(strings.TrimSpace(ext))
		if filepath.Ext(file) == ext {
			return true
		}
	}

	return false
}

// Gets the directory to store and locate images for a gallery.
func (gs *GalleryService) galleryDir(id int) string {
	imagesDir := gs.ImagesDir
	if imagesDir == "" {
		imagesDir = "images"
	}

	return filepath.Join(imagesDir, fmt.Sprintf("gallery-%d", id))
}

// Gets the supported image file extensions on the gallery service if set, otherwise uses defaults.
func (gs *GalleryService) supportedExt() []string {
	imagesExt := gs.ImagesExt
	if imagesExt == nil {
		imagesExt = []string{
			".png",
			".jpg",
			".jpeg",
			".gif",
		}
	}

	return imagesExt
}

// Gets the supported image file contentTypes on the gallery service if set, otherwise uses defaults.
func (gs *GalleryService) supportedType() []string {
	imagesType := gs.ImagesType
	if imagesType == nil {
		imagesType = []string{
			"image/png",
			"image/jpeg",
			"image/gif",
		}
	}

	return imagesType
}

func (gs *GalleryService) Create(title string, userID int) (*Gallery, error) {
	gallery := Gallery{
		UserID: userID,
		Title: title,
		Published: false,
	}

	row := gs.DB.QueryRow(`
		INSERT INTO galleries (user_id, title, published)
		VALUES ($1, $2, $3) RETURNING id;
	`, userID, title, false)
	err := row.Scan(&gallery.ID)
	if err != nil {
		return nil, fmt.Errorf("create gallery: %w", err)
	}

	return &gallery, nil
}

func (gs *GalleryService) GetById(id int) (*Gallery, error) {
	gallery := Gallery{
		ID: id,
	}

	row := gs.DB.QueryRow(`
		SELECT user_id, title, published
		FROM galleries WHERE id = $1
	`, id)
	err := row.Scan(&gallery.UserID, &gallery.Title, &gallery.Published)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get gallery: %w", err)
	}

	return &gallery, nil
}

func (gs *GalleryService) GetAllByUserID(userID int) ([]Gallery, error) {
	rows, err := gs.DB.Query(`
		SELECT id, title, published
		FROM galleries WHERE user_id = $1
	`, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get galleries by user: %w", err)
	}

	var galleries []Gallery
	for rows.Next() {
		gallery := Gallery{
			UserID: userID,
		}
		rows.Scan(&gallery.ID, &gallery.Title, &gallery.Published)
		galleries = append(galleries, gallery)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("get galleries by user: %w", err)
	}

	return galleries, nil
}

func (gs *GalleryService) Update(gallery *Gallery) error {
	_, err := gs.DB.Exec(`
		UPDATE galleries
		SET title = $2, published = $3
		WHERE id = $1
	`, gallery.ID, gallery.Title, gallery.Published)
	if err != nil {
		return fmt.Errorf("update gallery: %w", err)
	}

	return nil
}

func (gs *GalleryService) Delete(id int) error {
	_, err := gs.DB.Exec(`
		DELETE FROM galleries
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("delete gallery: %w", err)
	}

	return nil
}

func (gs *GalleryService) Images(galleryID int) ([]Image, error) {
	globPattern := filepath.Join(gs.galleryDir(galleryID), "*")
	allFiles, err := filepath.Glob(globPattern)
	if err != nil {
		return nil, fmt.Errorf("get images by gallery: %w", err)
	}

	var images []Image
	for _, file := range allFiles {
		if hasExtension(file, gs.supportedExt()...) {
			images = append(images, Image{
				GalleryID: galleryID,
				Path: file,
				Filename: filepath.Base(file),
			})
		}
	}

	return images, nil
}

func (gs *GalleryService) Image(galleryID int, filename string) (*Image, error) {
	imagePath := filepath.Join(gs.galleryDir(galleryID), filename)

	// Check if image exists
	_, err := os.Stat(imagePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get image: %w", err)
	}
	
	return &Image{
		GalleryID: galleryID,
		Path: imagePath,
		Filename: filepath.Base(imagePath),
	}, nil
}

func (gs *GalleryService) CreateImage(galleryID int, filename string, contents io.ReadSeeker) error {
	// Check file content type and extension is valid
	err := checkContentType(contents, gs.supportedType())
	if err != nil {
		return fmt.Errorf("create image %v: %w", filename, err)
	}
	err = checkExtension(filename, gs.supportedExt())
	if err != nil {
		return fmt.Errorf("create image %v: %w", filename, err)
	}

	// Create gallery images directory
	galleryDir := gs.galleryDir(galleryID)
	err = os.MkdirAll(galleryDir, 0755)
	if err != nil {
		return fmt.Errorf("create gallery-%d images directory: %w", galleryID, err)
	}

	// Create image file
	imagePath := filepath.Join(galleryDir, filename)
	dst, err := os.Create(imagePath)
	if err != nil {
		return fmt.Errorf("create image file: %w", err)
	}
	defer dst.Close()

	// Copy contents to image file
	_, err = io.Copy(dst, contents)
	if err != nil {
		return fmt.Errorf("copy contents to image file: %w", err)
	}

	return nil
}

func (gs *GalleryService) DeleteImage(galleryID int, filename string) error {
	image, err := gs.Image(galleryID, filename)
	if err != nil {
		return fmt.Errorf("delete image: %w", err)
	}

	err = os.Remove(image.Path)
	if err != nil {
		return fmt.Errorf("delete image: %w", err)
	}

	return nil
}
