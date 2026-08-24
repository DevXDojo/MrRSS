package service

import (
	"MrRSS/internal/database"
)

// settingsService implements SettingsService interface
type settingsService struct {
	db *database.DB
}

// NewSettingsService creates a new settings service
func NewSettingsService(db *database.DB) SettingsService {
	return &settingsService{db: db}
}

// Get retrieves a setting value
func (s *settingsService) Get(key string) (string, error) {
	return s.db.GetSetting(key)
}

// Set sets a setting value
func (s *settingsService) Set(key, value string) error {
	return s.db.SetSetting(key, value)
}

// GetEncrypted retrieves an encrypted setting value
func (s *settingsService) GetEncrypted(key string) (string, error) {
	return s.db.GetEncryptedSetting(key)
}

// SetEncrypted sets an encrypted setting value
func (s *settingsService) SetEncrypted(key, value string) error {
	return s.db.SetEncryptedSetting(key, value)
}

// GetAll retrieves all settings
func (s *settingsService) GetAll() (map[string]string, error) {
	// This is a simplified implementation
	// In a real scenario, we'd need to query all settings from the database
	return map[string]string{}, nil
}

// SaveAll saves multiple settings
func (s *settingsService) SaveAll(settings map[string]string) error {
	// This is a simplified implementation
	// In a real scenario, we'd need to batch save all settings
	for key, value := range settings {
		if err := s.db.SetSetting(key, value); err != nil {
			return err
		}
	}
	return nil
}
