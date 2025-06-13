package models

import (
	"database/sql"
	"errors"
	"fmt"
)

type Gallery struct {
	ID int
	UserID int
	Title string
	Published bool
}

type GalleryService struct {
	DB *sql.DB
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
