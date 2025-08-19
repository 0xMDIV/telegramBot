package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
}

type PendingUser struct {
	UserID     int64
	ChatID     int64
	CaptchaKey string
	ExpiresAt  time.Time
	Attempts   int
}

type MutedUser struct {
	UserID int64
	ChatID int64
	Until  time.Time
}

func NewDB(filepath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return db, nil
}

func (db *DB) createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS pending_users (
			user_id INTEGER,
			chat_id INTEGER,
			captcha_key TEXT,
			expires_at DATETIME,
			attempts INTEGER DEFAULT 0,
			PRIMARY KEY (user_id, chat_id)
		)`,
		`CREATE TABLE IF NOT EXISTS muted_users (
			user_id INTEGER,
			chat_id INTEGER,
			until DATETIME,
			PRIMARY KEY (user_id, chat_id)
		)`,
		`CREATE TABLE IF NOT EXISTS group_settings (
			chat_id INTEGER PRIMARY KEY,
			admin_ids TEXT,
			captcha_enabled BOOLEAN DEFAULT 1
		)`,
		`CREATE TABLE IF NOT EXISTS welcome_messages (
			user_id INTEGER,
			chat_id INTEGER,
			message_id INTEGER,
			PRIMARY KEY (user_id, chat_id)
		)`,
	}

	for _, query := range queries {
		if _, err := db.conn.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) AddPendingUser(user PendingUser) error {
	query := `INSERT OR REPLACE INTO pending_users (user_id, chat_id, captcha_key, expires_at, attempts) 
			  VALUES (?, ?, ?, ?, ?)`
	_, err := db.conn.Exec(query, user.UserID, user.ChatID, user.CaptchaKey, user.ExpiresAt, user.Attempts)
	return err
}

func (db *DB) GetPendingUser(userID, chatID int64) (*PendingUser, error) {
	query := `SELECT user_id, chat_id, captcha_key, expires_at, attempts FROM pending_users 
			  WHERE user_id = ? AND chat_id = ?`

	var user PendingUser
	err := db.conn.QueryRow(query, userID, chatID).Scan(
		&user.UserID, &user.ChatID, &user.CaptchaKey, &user.ExpiresAt, &user.Attempts,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *DB) RemovePendingUser(userID, chatID int64) error {
	query := `DELETE FROM pending_users WHERE user_id = ? AND chat_id = ?`
	_, err := db.conn.Exec(query, userID, chatID)
	return err
}

func (db *DB) IncrementAttempts(userID, chatID int64) error {
	query := `UPDATE pending_users SET attempts = attempts + 1 WHERE user_id = ? AND chat_id = ?`
	_, err := db.conn.Exec(query, userID, chatID)
	return err
}

func (db *DB) AddMutedUser(muted MutedUser) error {
	query := `INSERT OR REPLACE INTO muted_users (user_id, chat_id, until) VALUES (?, ?, ?)`
	_, err := db.conn.Exec(query, muted.UserID, muted.ChatID, muted.Until)
	return err
}

func (db *DB) IsUserMuted(userID, chatID int64) (bool, error) {
	query := `SELECT until FROM muted_users WHERE user_id = ? AND chat_id = ?`

	var until time.Time
	err := db.conn.QueryRow(query, userID, chatID).Scan(&until)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if time.Now().After(until) {
		db.RemoveMutedUser(userID, chatID)
		return false, nil
	}

	return true, nil
}

func (db *DB) RemoveMutedUser(userID, chatID int64) error {
	query := `DELETE FROM muted_users WHERE user_id = ? AND chat_id = ?`
	_, err := db.conn.Exec(query, userID, chatID)
	return err
}

func (db *DB) CleanExpiredUsers() error {
	query := `DELETE FROM pending_users WHERE expires_at < ?`
	_, err := db.conn.Exec(query, time.Now())
	return err
}

func (db *DB) SetWelcomeMessage(userID, chatID int64, messageID int) error {
	query := `INSERT OR REPLACE INTO welcome_messages (user_id, chat_id, message_id) VALUES (?, ?, ?)`
	_, err := db.conn.Exec(query, userID, chatID, messageID)
	return err
}

func (db *DB) GetWelcomeMessage(userID, chatID int64) (int, error) {
	query := `SELECT message_id FROM welcome_messages WHERE user_id = ? AND chat_id = ?`

	var messageID int
	err := db.conn.QueryRow(query, userID, chatID).Scan(&messageID)
	if err != nil {
		return 0, err
	}
	return messageID, nil
}

func (db *DB) RemoveWelcomeMessage(userID, chatID int64) error {
	query := `DELETE FROM welcome_messages WHERE user_id = ? AND chat_id = ?`
	_, err := db.conn.Exec(query, userID, chatID)
	return err
}

func (db *DB) Close() error {
	return db.conn.Close()
}
