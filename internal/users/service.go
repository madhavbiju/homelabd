package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/madhavbiju/homelabd/internal/database"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrUsernameTaken    = errors.New("username already taken")
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
)

type User struct {
	ID           string
	Username     string
	PasswordHash string `json:"-"` // Never expose
	Role         string
	CreatedAt    time.Time
}

type Service struct {
	db *database.Database
}

func NewService(db *database.Database) *Service {
	return &Service{db: db}
}

// CreateUser hashes the password and creates a new user.
func (s *Service) CreateUser(ctx context.Context, username, password, role string) (*User, error) {
	if len(password) < 8 {
		return nil, ErrPasswordTooShort
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &User{
		ID:           uuid.NewString(),
		Username:     username,
		PasswordHash: string(hash),
		Role:         role,
	}

	query := `INSERT INTO users (id, username, password_hash, role) VALUES (?, ?, ?, ?)`
	_, err = s.db.DB.ExecContext(ctx, query, user.ID, user.Username, user.PasswordHash, user.Role)
	if err != nil {
		// In a real app we'd check for sqlite unique constraint error here specifically
		return nil, fmt.Errorf("failed to create user (may already exist): %w", err)
	}

	return user, nil
}

// VerifyPassword checks if the username and password match.
func (s *Service) VerifyPassword(ctx context.Context, username, password string) (*User, error) {
	query := `SELECT id, username, password_hash, role, created_at FROM users WHERE username = ?`
	row := s.db.DB.QueryRowContext(ctx, query, username)

	var user User
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidPassword
	}

	return &user, nil
}

// Count returns the total number of users (useful for bootstrap checks).
func (s *Service) Count(ctx context.Context) (int, error) {
	var count int
	err := s.db.DB.QueryRowContext(ctx, "SELECT COUNT(id) FROM users").Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
