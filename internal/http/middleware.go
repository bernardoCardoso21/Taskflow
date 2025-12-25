package http

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type ctxKey string

const (
	ctxRequestID ctxKey = "request_id"
	ctxUserID    ctxKey = "user_id"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := make([]byte, 8)
		_, _ = rand.Read(b)
		id := hex.EncodeToString(b)
		ctx := context.WithValue(r.Context(), ctxRequestID, id)
		w.Header().Set("X-Request-Id", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AuthJWT(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if h == "" || !strings.HasPrefix(h, "Bearer ") {
				WriteError(w, 401, "UNAUTHORIZED", "missing bearer token", nil)
				return
			}
			raw := strings.TrimPrefix(h, "Bearer ")

			tok, err := jwt.Parse(raw, func(token *jwt.Token) (any, error) {
				return []byte(secret), nil
			})
			if err != nil || !tok.Valid {
				WriteError(w, 401, "UNAUTHORIZED", "invalid token", nil)
				return
			}
			claims, ok := tok.Claims.(jwt.MapClaims)
			if !ok {
				WriteError(w, 401, "UNAUTHORIZED", "invalid token claims", nil)
				return
			}
			uid, _ := claims["sub"].(string)
			if uid == "" {
				WriteError(w, 401, "UNAUTHORIZED", "missing subject", nil)
				return
			}

			ctx := context.WithValue(r.Context(), ctxUserID, uid)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserID(ctx context.Context) (string, bool) {
	v := ctx.Value(ctxUserID)
	s, ok := v.(string)
	return s, ok
}
