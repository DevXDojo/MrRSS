package database

import "context"

// CompareAndSwapSetting prevents a settings window from overwriting a newer
// edit. Defaults are seeded at initialization, but missing keys are supported.
func (db *DB) CompareAndSwapSetting(ctx context.Context, key, old, next string) (bool, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	if old == "" {
		if _, err = tx.ExecContext(ctx, "INSERT OR IGNORE INTO settings(key,value) VALUES(?,?)", key, old); err != nil {
			return false, err
		}
	}
	result, err := tx.ExecContext(ctx, "UPDATE settings SET value=? WHERE key=? AND value=?", next, key, old)
	if err != nil {
		return false, err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if err = tx.Commit(); err != nil {
		return false, err
	}
	return count == 1, nil
}
