package database

import (
	"auth_service/logger"
	"auth_service/tokens"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type ITokenRepository interface {
	SaveRefreshToken(ctx context.Context, t *tokens.RefreshToken) error
	GetRefreshToken(ctx context.Context, tokenVal string) (*tokens.RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, id uuid.UUID) error
}

type TokenRepository struct {
	db *sql.DB
}

func NewTokenRepository(conn *dbConnection) ITokenRepository {
	return &TokenRepository{
		db: conn.db,
	}
}

func IsUniqueConstraintError(err error) bool {
    var pqErr *pq.Error
    if errors.As(err, &pqErr) {
        return pqErr.Code == "23505"
    }
    return false
}

func (r *TokenRepository) SaveRefreshToken(ctx context.Context, t *tokens.RefreshToken) error {
	query := `INSERT INTO refresh_token (id, userid, refreshtoken, ip, createdat, expiresat, useragent) VALUES ($1, $2, $3, $4, $5, $6, $7);`
	if _, err := r.db.ExecContext(ctx, query, t.Id, t.UserId, t.HashToken(), t.ClientIp, t.CreatedAt, t.ExpiresAt, t.UserAgent); err != nil {
		logger.Err.Println("refresh token saving failed -", err)
		return err
	}

	return nil
}

func (r *TokenRepository) GetRefreshToken(ctx context.Context, tokenVal string) (*tokens.RefreshToken, error) {
	found := tokens.RefreshToken{}
	query := `SELECT * FROM refresh_token WHERE expiresat > NOW()`
	rows, err := r.db.QueryContext(ctx, query)

	if err != nil {
		fmt.Println("refresh token search failed -", err)
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var t tokens.RefreshToken
		var thash []byte 

		if err := rows.Scan(&t.Id, &t.UserId, &thash, &t.ClientIp, &t.CreatedAt, &t.ExpiresAt, &t.UserAgent); err != nil {
			return nil, err
		}

		if err := bcrypt.CompareHashAndPassword(thash, []byte(tokenVal)); err == nil {
            found = t
			found.Token = tokenVal
            return &found, nil
        }
	}

	return nil, sql.ErrNoRows
}

func (r *TokenRepository) DeleteRefreshToken(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM refresh_token WHERE id = $1`
	if _, err := r.db.ExecContext(ctx, query, id); err != nil {
		fmt.Println("refresh token deletion failed -", err)
		return err
	}

	return nil
}
