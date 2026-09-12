package database

import (
	"MrRSS/internal/crypto"
	"context"
	"fmt"
	"log"
)

// GetSetting retrieves a setting value by key.
func (db *DB) GetSetting(key string) (string, error) {
	return db.GetSettingContext(context.Background(), key)
}

// GetSettingContext reads a setting with cancellation support.
func (db *DB) GetSettingContext(ctx context.Context, key string) (string, error) {
	select {
	case <-db.ready:
	case <-ctx.Done():
		return "", ctx.Err()
	}
	var value string
	err := db.QueryRowContext(ctx, "SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	if err != nil {
		return "", err
	}
	return value, nil
}

// SetSetting stores a setting value.
func (db *DB) SetSetting(key, value string) error {
	db.WaitForReady()
	_, err := db.Exec("INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)", key, value)
	return err
}

// GetEncryptedSetting retrieves and decrypts a sensitive setting value.
// If the value is not encrypted (plain text), it will be automatically encrypted
// and stored back to support migration from old versions.
func (db *DB) GetEncryptedSetting(key string) (string, error) {
	db.WaitForReady()

	// Get the stored value
	storedValue, err := db.GetSetting(key)
	if err != nil {
		return "", err
	}

	// Empty value - return as is
	if storedValue == "" {
		return "", nil
	}

	// Check if the value is already encrypted
	if crypto.IsEncrypted(storedValue) {
		// Decrypt and return
		decrypted, err := crypto.Decrypt(storedValue)
		if err != nil {
			return "", fmt.Errorf("failed to decrypt setting %s: %w", key, err)
		}
		return decrypted, nil
	}

	// Value is plain text - migrate it to encrypted format
	log.Printf("Migrating plain text setting to encrypted storage")

	// Encrypt the plain text value
	encrypted, err := crypto.Encrypt(storedValue)
	if err != nil {
		// If encryption fails, return an error to the caller
		log.Printf("Warning: Failed to encrypt setting during migration: %v", err)
		return "", fmt.Errorf("failed to encrypt setting during migration: %w", err)
	}

	// Store the encrypted value back
	if err := db.SetSetting(key, encrypted); err != nil {
		// If storage fails, return an error to the caller
		log.Printf("Warning: Failed to store encrypted setting: %v", err)
		return "", fmt.Errorf("failed to store encrypted setting: %w", err)
	}

	// Return the original plain text value
	return storedValue, nil
}

// SetEncryptedSetting encrypts and stores a sensitive setting value.
func (db *DB) SetEncryptedSetting(key, value string) error {
	db.WaitForReady()

	// Empty value - store as is
	if value == "" {
		return db.SetSetting(key, value)
	}

	// Encrypt the value
	encrypted, err := crypto.Encrypt(value)
	if err != nil {
		return fmt.Errorf("failed to encrypt setting %s: %w", key, err)
	}

	// Store the encrypted value
	return db.SetSetting(key, encrypted)
}
