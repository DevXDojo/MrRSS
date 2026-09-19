package database

import (
	"context"
	"database/sql"
	"time"

	"MrRSS/internal/models"
)

const notificationSchema = `
CREATE TABLE IF NOT EXISTS notification_state (
 state_key TEXT PRIMARY KEY, cursor INTEGER NOT NULL, last_run INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS notification_deliveries (
 id INTEGER PRIMARY KEY AUTOINCREMENT, state_key TEXT NOT NULL,
 channel_id TEXT NOT NULL, rule_id TEXT NOT NULL, label TEXT NOT NULL,
 body TEXT NOT NULL, status TEXT NOT NULL DEFAULT 'pending',
 attempts INTEGER NOT NULL DEFAULT 0, next_attempt INTEGER NOT NULL DEFAULT 0,
 error_code TEXT NOT NULL DEFAULT '', created_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_notification_pending ON notification_deliveries(status,next_attempt,id);
`

type NotificationArticle struct {
	models.Article
	Category string
}

type NotificationState struct {
	Cursor  int64
	LastRun time.Time
}

type NotificationDelivery struct {
	ID          int64  `json:"id"`
	StateKey    string `json:"-"`
	ChannelID   string `json:"channel_id"`
	RuleID      string `json:"rule_id"`
	Label       string `json:"label"`
	Body        string `json:"-"`
	Status      string `json:"status"`
	Attempts    int    `json:"attempts"`
	NextAttempt int64  `json:"next_attempt"`
	ErrorCode   string `json:"error_code"`
	CreatedAt   int64  `json:"created_at"`
}

// SaveNotificationConfig commits credentials and rule baselines together. Newly
// enabled rules start at the current article ID, never at the start of an archive.
func (db *DB) SaveNotificationConfig(ctx context.Context, encrypted string, active, reset map[string]bool, now time.Time) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, "INSERT OR REPLACE INTO settings(key,value) VALUES('notification_config',?)", encrypted); err != nil {
		return err
	}
	rows, err := tx.QueryContext(ctx, "SELECT state_key FROM notification_state")
	if err != nil {
		return err
	}
	var remove []string
	for rows.Next() {
		var key string
		if err = rows.Scan(&key); err != nil {
			rows.Close()
			return err
		}
		if !active[key] || reset[key] {
			remove = append(remove, key)
		}
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	for _, key := range remove {
		if _, err = tx.ExecContext(ctx, "DELETE FROM notification_state WHERE state_key=?", key); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, "UPDATE notification_deliveries SET status='cancelled',body='' WHERE state_key=? AND status='pending'", key); err != nil {
			return err
		}
	}
	for key := range active {
		if _, err = tx.ExecContext(ctx, `INSERT OR IGNORE INTO notification_state(state_key,cursor,last_run) SELECT ?,COALESCE(MAX(id),0),? FROM articles`, key, now.Unix()); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (db *DB) GetNotificationState(ctx context.Context, key string) (NotificationState, error) {
	var s NotificationState
	var last int64
	err := db.QueryRowContext(ctx, "SELECT cursor,last_run FROM notification_state WHERE state_key=?", key).Scan(&s.Cursor, &last)
	s.LastRun = time.Unix(last, 0)
	return s, err
}

// NotificationArticles reads a bounded batch. Preview reads newest first;
// scheduled evaluation processes new IDs in order, including backdated feeds.
func (db *DB) NotificationArticles(ctx context.Context, after int64, preview bool) ([]NotificationArticle, error) {
	order := "ASC"
	if preview {
		order = "DESC"
	}
	rows, err := db.QueryContext(ctx, `SELECT a.id,a.feed_id,COALESCE(a.title,''),COALESCE(a.url,''),
 COALESCE(a.summary,''),COALESCE(a.original_summary,''),COALESCE(a.author,''),
 a.is_read,a.is_favorite,a.is_read_later,a.is_hidden,COALESCE(f.title,''),COALESCE(f.category,'')
 FROM articles a JOIN feeds f ON f.id=a.feed_id WHERE a.id>? ORDER BY a.id `+order+` LIMIT 500`, after)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []NotificationArticle{}
	for rows.Next() {
		var a NotificationArticle
		if err = rows.Scan(&a.ID, &a.FeedID, &a.Title, &a.URL, &a.Summary, &a.OriginalSummary, &a.Author, &a.IsRead, &a.IsFavorite, &a.IsReadLater, &a.IsHidden, &a.FeedTitle, &a.Category); err != nil {
			return nil, err
		}
		items = append(items, a)
	}
	return items, rows.Err()
}

// QueueNotification advances the scan only in the same transaction as its
// messages. A crash cannot silently lose a batch between scanning and queuing.
func (db *DB) QueueNotification(ctx context.Context, key, channelID, ruleID, label string, cursor int64, now time.Time, bodies []string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, body := range bodies {
		if _, err = tx.ExecContext(ctx, `INSERT INTO notification_deliveries(state_key,channel_id,rule_id,label,body,created_at) VALUES(?,?,?,?,?,?)`, key, channelID, ruleID, label, body, now.Unix()); err != nil {
			return err
		}
	}
	if _, err = tx.ExecContext(ctx, "UPDATE notification_state SET cursor=?,last_run=? WHERE state_key=?", cursor, now.Unix(), key); err != nil {
		return err
	}
	return tx.Commit()
}

func (db *DB) NotificationDeliveries(ctx context.Context, pending bool, now time.Time) ([]NotificationDelivery, error) {
	query := `SELECT id,state_key,channel_id,rule_id,label,body,status,attempts,next_attempt,error_code,created_at FROM notification_deliveries`
	var rows *sql.Rows
	var err error
	if pending {
		// Preserve chunk ordering and avoid bursts to the same channel. Later
		// messages wait while an earlier message is backing off.
		rows, err = db.QueryContext(ctx, query+` d WHERE status='pending' AND next_attempt<=? AND NOT EXISTS (SELECT 1 FROM notification_deliveries earlier WHERE earlier.channel_id=d.channel_id AND earlier.status='pending' AND earlier.id<d.id) ORDER BY id LIMIT 20`, now.Unix())
	} else {
		rows, err = db.QueryContext(ctx, query+` ORDER BY id DESC LIMIT 100`)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []NotificationDelivery{}
	for rows.Next() {
		var d NotificationDelivery
		if err = rows.Scan(&d.ID, &d.StateKey, &d.ChannelID, &d.RuleID, &d.Label, &d.Body, &d.Status, &d.Attempts, &d.NextAttempt, &d.ErrorCode, &d.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, d)
	}
	return items, rows.Err()
}

func (db *DB) CompleteNotification(ctx context.Context, id int64, status, code string, next time.Time) error {
	_, err := db.ExecContext(ctx, `UPDATE notification_deliveries SET status=?,error_code=?,attempts=attempts+1,next_attempt=?,body=CASE WHEN ?='pending' THEN body ELSE '' END WHERE id=?`, status, code, next.Unix(), status, id)
	return err
}

func (db *DB) PruneNotifications(ctx context.Context, now time.Time) error {
	_, err := db.ExecContext(ctx, `DELETE FROM notification_deliveries WHERE (status!='pending' AND created_at<?) OR id IN (SELECT id FROM notification_deliveries WHERE status!='pending' ORDER BY id DESC LIMIT -1 OFFSET 1000)`, now.AddDate(0, 0, -30).Unix())
	return err
}

func (db *DB) NotificationPendingCount(ctx context.Context, key string) (int, error) {
	var n int
	err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM notification_deliveries WHERE state_key=? AND status='pending'`, key).Scan(&n)
	return n, err
}
