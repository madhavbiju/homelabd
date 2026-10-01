package ctxutil

import (
	"context"

	"github.com/madhavbiju/homelabd/internal/auth"
)

type contextKey string

const (
	RequestIDKey   contextKey = "request_id"
	TokenRecordKey contextKey = "token_record"
)

func GetRequestID(ctx context.Context) string {
	if reqID, ok := ctx.Value(RequestIDKey).(string); ok {
		return reqID
	}
	return ""
}

func GetTokenRecord(ctx context.Context) *auth.TokenRecord {
	if record, ok := ctx.Value(TokenRecordKey).(*auth.TokenRecord); ok {
		return record
	}
	return nil
}
