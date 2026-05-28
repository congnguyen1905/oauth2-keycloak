package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"auth-service/pkg/models"
	
	_ "github.com/lib/pq" // PostgreSQL driver
	_ "github.com/go-sql-driver/mysql" // MySQL driver
)

// DatabaseStorage implements Storage using a relational database
type DatabaseStorage struct {
	db     *sql.DB
	driver string
	prefix string
}

// NewDatabaseStorage creates a new database storage instance
func NewDatabaseStorage(driver, dsn string, maxIdle, maxOpen int) (*DatabaseStorage, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	
	db.SetMaxIdleConns(maxIdle)
	db.SetMaxOpenConns(maxOpen)
	
	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	
	// Create sessions table if not exists
	if err := createSessionsTable(db, driver); err != nil {
		return nil, fmt.Errorf("failed to create sessions table: %w", err)
	}
	
	return &DatabaseStorage{
		db:     db,
		driver: driver,
		prefix: "",
	}, nil
}

func createSessionsTable(db *sql.DB, driver string) error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS sessions (
		session_id VARCHAR(255) PRIMARY KEY,
		user_id VARCHAR(255) NOT NULL,
		username VARCHAR(255) NOT NULL,
		email VARCHAR(255),
		roles JSONB,
		access_token TEXT,
		expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		claims JSONB
	);
	
	CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);
	`
	
	if driver == "mysql" {
		createTableSQL = `
		CREATE TABLE IF NOT EXISTS sessions (
			session_id VARCHAR(255) PRIMARY KEY,
			user_id VARCHAR(255) NOT NULL,
			username VARCHAR(255) NOT NULL,
			email VARCHAR(255),
			roles JSON,
			access_token TEXT,
			expires_at DATETIME NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			claims JSON,
			INDEX idx_expires_at (expires_at)
		);
		`
	}
	
	_, err := db.Exec(createTableSQL)
	return err
}

// Save saves a session to the database
func (d *DatabaseStorage) Save(ctx context.Context, session *models.UserSession, ttl time.Duration) error {
	rolesJSON, _ := json.Marshal(session.Roles)
	claimsJSON, _ := json.Marshal(session.Claims)
	
	insertSQL := `
	INSERT INTO sessions (session_id, user_id, username, email, roles, access_token, expires_at, claims)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	ON CONFLICT (session_id) DO UPDATE SET
		user_id = EXCLUDED.user_id,
		username = EXCLUDED.username,
		email = EXCLUDED.email,
		roles = EXCLUDED.roles,
		access_token = EXCLUDED.access_token,
		expires_at = EXCLUDED.expires_at,
		claims = EXCLUDED.claims
	`
	
	_, err := d.db.ExecContext(ctx, insertSQL,
		session.SessionID,
		session.UserID,
		session.Username,
		session.Email,
		rolesJSON,
		session.AccessToken,
		session.ExpiresAt,
		claimsJSON,
	)
	
	return err
}

// Get retrieves a session from the database
func (d *DatabaseStorage) Get(ctx context.Context, sessionID string) (*models.UserSession, error) {
	query := `
	SELECT session_id, user_id, username, email, roles, access_token, expires_at, created_at, claims
	FROM sessions
	WHERE session_id = $1 AND expires_at > NOW()
	`
	
	row := d.db.QueryRowContext(ctx, query, sessionID)
	
	var session models.UserSession
	var rolesJSON, claimsJSON []byte
	
	err := row.Scan(
		&session.SessionID,
		&session.UserID,
		&session.Username,
		&session.Email,
		&rolesJSON,
		&session.AccessToken,
		&session.ExpiresAt,
		&session.CreatedAt,
		&claimsJSON,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	json.Unmarshal(rolesJSON, &session.Roles)
	json.Unmarshal(claimsJSON, &session.Claims)
	
	return &session, nil
}

// Delete removes a session from the database
func (d *DatabaseStorage) Delete(ctx context.Context, sessionID string) error {
	_, err := d.db.ExecContext(ctx, "DELETE FROM sessions WHERE session_id = $1", sessionID)
	return err
}

// Exists checks if a session exists in the database
func (d *DatabaseStorage) Exists(ctx context.Context, sessionID string) (bool, error) {
	var exists bool
	err := d.db.QueryRowContext(ctx, 
		"SELECT EXISTS(SELECT 1 FROM sessions WHERE session_id = $1 AND expires_at > NOW())",
		sessionID,
	).Scan(&exists)
	
	return exists, err
}

// UpdateTTL updates the TTL of a session (extends expiration)
func (d *DatabaseStorage) UpdateTTL(ctx context.Context, sessionID string, ttl time.Duration) error {
	newExpires := time.Now().Add(ttl)
	
	_, err := d.db.ExecContext(ctx,
		"UPDATE sessions SET expires_at = $1 WHERE session_id = $2",
		newExpires, sessionID,
	)
	
	return err
}

// Close closes the database connection
func (d *DatabaseStorage) Close() error {
	return d.db.Close()
}

// CleanupExpiredSessions removes expired sessions from the database
func (d *DatabaseStorage) CleanupExpiredSessions(ctx context.Context) (int64, error) {
	result, err := d.db.ExecContext(ctx, "DELETE FROM sessions WHERE expires_at < NOW()")
	if err != nil {
		return 0, err
	}
	
	return result.RowsAffected()
}
