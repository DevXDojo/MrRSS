package database

import (
	"database/sql"
	"time"

	"MrRSS/internal/models"
)

// GetSavedFilters retrieves all saved filters ordered by position
func (db *DB) GetSavedFilters() ([]models.SavedFilter, error) {
	db.WaitForReady()

	rows, err := db.Query(`
		SELECT id, name, conditions, position, created_at, updated_at
		FROM saved_filters
		ORDER BY position ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var filters []models.SavedFilter
	for rows.Next() {
		var f models.SavedFilter
		var createdAt, updatedAt string

		if err := rows.Scan(&f.ID, &f.Name, &f.Conditions, &f.Position, &createdAt, &updatedAt); err != nil {
			return nil, err
		}

		// Parse timestamps
		f.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		f.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

		filters = append(filters, f)
	}

	return filters, nil
}

// AddSavedFilter creates a new saved filter
func (db *DB) AddSavedFilter(filter *models.SavedFilter) (int64, error) {
	db.WaitForReady()

	// Get next position if not specified
	if filter.Position == 0 {
		nextPos, err := db.GetNextSavedFilterPosition()
		if err != nil {
			return 0, err
		}
		filter.Position = nextPos
	}

	result, err := db.Exec(`
		INSERT INTO saved_filters (name, conditions, position, created_at, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, filter.Name, filter.Conditions, filter.Position)

	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Query the newly created filter to get timestamps
	var createdAt, updatedAt string
	err = db.QueryRow(`
		SELECT created_at, updated_at FROM saved_filters WHERE id = ?
	`, id).Scan(&createdAt, &updatedAt)
	if err != nil {
		// If we can't get timestamps, still return the ID
		return id, nil
	}

	// Parse timestamps
	filter.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	filter.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	return id, nil
}

// UpdateSavedFilter updates an existing saved filter
func (db *DB) UpdateSavedFilter(filter *models.SavedFilter) error {
	db.WaitForReady()

	_, err := db.Exec(`
		UPDATE saved_filters
		SET name = ?, conditions = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, filter.Name, filter.Conditions, filter.ID)

	return err
}

// DeleteSavedFilter removes a saved filter
func (db *DB) DeleteSavedFilter(id int64) error {
	db.WaitForReady()

	_, err := db.Exec(`DELETE FROM saved_filters WHERE id = ?`, id)
	return err
}

// ReorderSavedFilters updates filter positions in bulk
func (db *DB) ReorderSavedFilters(filters []models.SavedFilter) error {
	db.WaitForReady()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`UPDATE saved_filters SET position = ? WHERE id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, filter := range filters {
		if _, err := stmt.Exec(filter.Position, filter.ID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetNextSavedFilterPosition returns the next position value
func (db *DB) GetNextSavedFilterPosition() (int, error) {
	db.WaitForReady()

	var maxPos int
	err := db.QueryRow(`SELECT COALESCE(MAX(position), 0) FROM saved_filters`).Scan(&maxPos)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	return maxPos + 1, nil
}
