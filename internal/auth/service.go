package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/madhavbiju/homelabd/internal/database"
)

var (
	ErrTokenInvalid = errors.New("invalid or revoked token")
	ErrTokenExpired = errors.New("token expired")
)

type TokenRecord struct {
	ID          string
	Description string
	TokenHash   string
	Permissions []string
	CreatedAt   time.Time
	ExpiresAt   *time.Time
	Revoked     bool
}

type Service struct {
	db *database.Database
}

func NewService(db *database.Database) *Service {
	return &Service{db: db}
}

// CreateToken generates a new token and stores its hash in the database.
func (s *Service) CreateToken(ctx context.Context, description string, permissions []string, expiresAt *time.Time) (string, *TokenRecord, error) {
	plainToken, tokenHash, err := GenerateToken()
	if err != nil {
		return "", nil, err
	}

	id := uuid.NewString()
	permsJSON, err := json.Marshal(permissions)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal permissions: %w", err)
	}

	query := `
		INSERT INTO tokens (id, description, token_hash, permissions, created_at, expires_at, revoked)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	
	now := time.Now().UTC()
	var expiresAtVal interface{}
	if expiresAt != nil {
		expiresAtVal = expiresAt.UTC()
	}

	_, err = s.db.DB.ExecContext(ctx, query, id, description, tokenHash, string(permsJSON), now, expiresAtVal, false)
	if err != nil {
		return "", nil, fmt.Errorf("failed to insert token: %w", err)
	}

	record := &TokenRecord{
		ID:          id,
		Description: description,
		TokenHash:   tokenHash,
		Permissions: permissions,
		CreatedAt:   now,
		ExpiresAt:   expiresAt,
		Revoked:     false,
	}

	return plainToken, record, nil
}

// ValidateToken checks if a plain token is valid and returns its record.
func (s *Service) ValidateToken(ctx context.Context, plainToken string) (*TokenRecord, error) {
	hash := HashToken(plainToken)

	query := `
		SELECT id, description, token_hash, permissions, created_at, expires_at, revoked
		FROM tokens
		WHERE token_hash = ?
	`

	var record TokenRecord
	var permsJSON string
	var expiresAt sql.NullTime

	err := s.db.DB.QueryRowContext(ctx, query, hash).Scan(
		&record.ID,
		&record.Description,
		&record.TokenHash,
		&permsJSON,
		&record.CreatedAt,
		&expiresAt,
		&record.Revoked,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTokenInvalid
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	if record.Revoked {
		return nil, ErrTokenInvalid
	}

	if expiresAt.Valid {
		if time.Now().UTC().After(expiresAt.Time) {
			return nil, ErrTokenExpired
		}
		record.ExpiresAt = &expiresAt.Time
	}

	if err := json.Unmarshal([]byte(permsJSON), &record.Permissions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
	}

	return &record, nil
}

// RevokeToken marks a token as revoked.
func (s *Service) RevokeToken(ctx context.Context, id string) error {
	query := `UPDATE tokens SET revoked = TRUE WHERE id = ?`
	res, err := s.db.DB.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}
	
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("token not found")
	}
	
	return nil
}
