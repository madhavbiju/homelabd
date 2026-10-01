package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/madhavbiju/homelabd/internal/database"
)

type Event struct {
	TokenID      string
	Action       string
	Target       string
	Parameters   map[string]interface{}
	Result       string
	ErrorMessage string
	ClientIP     string
}

type Record struct {
	ID           int64                  `json:"id"`
	Timestamp    time.Time              `json:"timestamp"`
	TokenID      string                 `json:"token_id"`
	Action       string                 `json:"action"`
	Target       string                 `json:"target"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
	Result       string                 `json:"result"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	ClientIP     string                 `json:"client_ip"`
}

type Service struct {
	db *database.Database
}

func NewService(db *database.Database) *Service {
	return &Service{db: db}
}

// Log records an audit event to the database.
func (s *Service) Log(ctx context.Context, event Event) error {
	var paramsJSON string
	if event.Parameters != nil {
		b, err := json.Marshal(event.Parameters)
		if err == nil {
			paramsJSON = string(b)
		}
	}

	query := `
		INSERT INTO audit_logs (token_id, action, target, parameters, result, error_message, client_ip)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	
	_, err := s.db.DB.ExecContext(ctx, query,
		event.TokenID,
		event.Action,
		event.Target,
		paramsJSON,
		event.Result,
		event.ErrorMessage,
		event.ClientIP,
	)
	
	
	if err != nil {
		return fmt.Errorf("failed to write audit log: %w", err)
	}
	
	return nil
}

// List retrieves paginated audit logs.
func (s *Service) List(ctx context.Context, limit, offset int) ([]Record, error) {
	query := `
		SELECT id, timestamp, token_id, action, target, parameters, result, error_message, client_ip
		FROM audit_logs
		ORDER BY timestamp DESC
		LIMIT ? OFFSET ?
	`
	
	rows, err := s.db.DB.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var rec Record
		var paramsJSON string
		err := rows.Scan(
			&rec.ID,
			&rec.Timestamp,
			&rec.TokenID,
			&rec.Action,
			&rec.Target,
			&paramsJSON,
			&rec.Result,
			&rec.ErrorMessage,
			&rec.ClientIP,
		)
		if err != nil {
			return nil, err
		}
		if paramsJSON != "" {
			json.Unmarshal([]byte(paramsJSON), &rec.Parameters)
		}
		records = append(records, rec)
	}

	return records, nil
}
