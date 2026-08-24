package database

import (
	"database/sql"
	"fmt"
	"time"
)

// ChatSession represents a chat session for an article
type ChatSession struct {
	ID           int64     `json:"id"`
	ArticleID    int64     `json:"article_id"`
	Title        string    `json:"title"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	MessageCount int       `json:"message_count"`
}

// ChatMessage represents a message in a chat session
type ChatMessage struct {
	ID        int64     `json:"id"`
	SessionID int64     `json:"session_id"`
	Role      string    `json:"role"` // "user" or "assistant"
	Content   string    `json:"content"`
	Thinking  string    `json:"thinking,omitempty"` // AI thinking process (optional)
	CreatedAt time.Time `json:"created_at"`
}

// CreateChatSession creates a new chat session for an article
func (db *DB) CreateChatSession(articleID int64, title string) (int64, error) {
	result, err := db.Exec(
		`INSERT INTO chat_sessions (article_id, title, created_at, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		articleID, title,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create chat session: %w", err)
	}
	return result.LastInsertId()
}

// GetChatSession retrieves a chat session by ID
func (db *DB) GetChatSession(sessionID int64) (*ChatSession, error) {
	var session ChatSession
	err := db.QueryRow(`
		SELECT id, article_id, title, created_at, updated_at,
		       (SELECT COUNT(*) FROM chat_messages WHERE session_id = chat_sessions.id) as message_count
		FROM chat_sessions
		WHERE id = ?
	`, sessionID).Scan(
		&session.ID, &session.ArticleID, &session.Title,
		&session.CreatedAt, &session.UpdatedAt, &session.MessageCount,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get chat session: %w", err)
	}
	return &session, nil
}

// GetChatSessionsByArticle retrieves all chat sessions for an article, ordered by updated_at desc
func (db *DB) GetChatSessionsByArticle(articleID int64) ([]ChatSession, error) {
	rows, err := db.Query(`
		SELECT id, article_id, title, created_at, updated_at,
		       (SELECT COUNT(*) FROM chat_messages WHERE session_id = chat_sessions.id) as message_count
		FROM chat_sessions
		WHERE article_id = ?
		ORDER BY updated_at DESC
	`, articleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat sessions: %w", err)
	}
	defer rows.Close()

	sessions := make([]ChatSession, 0)
	for rows.Next() {
		var session ChatSession
		err := rows.Scan(
			&session.ID, &session.ArticleID, &session.Title,
			&session.CreatedAt, &session.UpdatedAt, &session.MessageCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chat session: %w", err)
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// UpdateChatSessionTitle updates the title of a chat session
func (db *DB) UpdateChatSessionTitle(sessionID int64, title string) error {
	_, err := db.Exec(
		`UPDATE chat_sessions SET title = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		title, sessionID,
	)
	if err != nil {
		return fmt.Errorf("failed to update chat session title: %w", err)
	}
	return nil
}

// UpdateChatSessionTimestamp updates the updated_at timestamp of a chat session
func (db *DB) UpdateChatSessionTimestamp(sessionID int64) error {
	_, err := db.Exec(
		`UPDATE chat_sessions SET updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		sessionID,
	)
	if err != nil {
		return fmt.Errorf("failed to update chat session timestamp: %w", err)
	}
	return nil
}

// DeleteChatSession deletes a chat session and all its messages
func (db *DB) DeleteChatSession(sessionID int64) error {
	_, err := db.Exec(`DELETE FROM chat_messages WHERE session_id = ?`, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete chat messages: %w", err)
	}
	_, err = db.Exec(`DELETE FROM chat_sessions WHERE id = ?`, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete chat session: %w", err)
	}
	return nil
}

// CreateChatMessage creates a new chat message in a session
func (db *DB) CreateChatMessage(sessionID int64, role, content, thinking string) (int64, error) {
	result, err := db.Exec(
		`INSERT INTO chat_messages (session_id, role, content, thinking, created_at) VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		sessionID, role, content, thinking,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create chat message: %w", err)
	}

	// Update session timestamp
	_ = db.UpdateChatSessionTimestamp(sessionID)

	return result.LastInsertId()
}

// GetChatMessages retrieves all messages for a session, ordered by created_at asc
func (db *DB) GetChatMessages(sessionID int64) ([]ChatMessage, error) {
	rows, err := db.Query(`
		SELECT id, session_id, role, content, thinking, created_at
		FROM chat_messages
		WHERE session_id = ?
		ORDER BY created_at ASC
	`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat messages: %w", err)
	}
	defer rows.Close()

	messages := make([]ChatMessage, 0)
	for rows.Next() {
		var msg ChatMessage
		var thinking sql.NullString
		err := rows.Scan(
			&msg.ID, &msg.SessionID, &msg.Role, &msg.Content,
			&thinking, &msg.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chat message: %w", err)
		}
		if thinking.Valid {
			msg.Thinking = thinking.String
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// DeleteChatMessage deletes a single chat message
func (db *DB) DeleteChatMessage(messageID int64) error {
	// Get session ID before deleting
	var sessionID int64
	err := db.QueryRow(`SELECT session_id FROM chat_messages WHERE id = ?`, messageID).Scan(&sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session ID: %w", err)
	}

	// Delete the message
	_, err = db.Exec(`DELETE FROM chat_messages WHERE id = ?`, messageID)
	if err != nil {
		return fmt.Errorf("failed to delete chat message: %w", err)
	}

	// Update session timestamp
	_ = db.UpdateChatSessionTimestamp(sessionID)

	return nil
}

// CleanupOldChatSessions removes chat sessions older than maxAgeDays
func (db *DB) CleanupOldChatSessions(maxAgeDays int) (int64, error) {
	// First delete messages
	_, err := db.Exec(
		`DELETE FROM chat_messages WHERE session_id IN (SELECT id FROM chat_sessions WHERE created_at < datetime('now', ?))`,
		fmt.Sprintf("-%d days", maxAgeDays),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old chat messages: %w", err)
	}

	// Then delete sessions
	result, err := db.Exec(
		`DELETE FROM chat_sessions WHERE created_at < datetime('now', ?)`,
		fmt.Sprintf("-%d days", maxAgeDays),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old chat sessions: %w", err)
	}

	return result.RowsAffected()
}

// DeleteAllChatSessions deletes all chat sessions and their messages
func (db *DB) DeleteAllChatSessions() (int64, error) {
	// First delete all messages
	_, err := db.Exec(`DELETE FROM chat_messages`)
	if err != nil {
		return 0, fmt.Errorf("failed to delete all chat messages: %w", err)
	}

	// Then delete all sessions
	result, err := db.Exec(`DELETE FROM chat_sessions`)
	if err != nil {
		return 0, fmt.Errorf("failed to delete all chat sessions: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get affected rows: %w", err)
	}

	return count, nil
}
