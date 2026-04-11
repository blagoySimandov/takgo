package sqlite

import (
	"context"
	"strings"
	"time"

	"github.com/blagoySimandov/takgo/internal/domain/auth"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type userModel struct {
	bun.BaseModel `bun:"table:users"`
	ID            string    `bun:"id,pk"`
	Username      string    `bun:"username,unique,notnull"`
	PasswordHash  string    `bun:"password_hash,notnull"`
	CreatedAt     time.Time `bun:"created_at,notnull"`
}

type UserRepo struct {
	db *bun.DB
}

func NewUserRepo(db *bun.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user *auth.User) error {
	_, err := r.db.NewInsert().Model(toModel(user)).Exec(ctx)
	if err != nil && strings.Contains(err.Error(), "UNIQUE constraint") {
		return auth.ErrUserExists
	}
	return err
}

func (r *UserRepo) FindByUsername(ctx context.Context, username string) (*auth.User, error) {
	m := &userModel{}
	err := r.db.NewSelect().Model(m).Where("username = ?", username).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return toDomain(m), nil
}

func toModel(u *auth.User) *userModel {
	return &userModel{
		// NOTE: SQLite has no native UUID type; stored as TEXT
		ID:           u.ID.String(),
		Username:     u.Username,
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt,
	}
}

func toDomain(m *userModel) *auth.User {
	return &auth.User{
		// NOTE: SQLite has no native UUID type; parsed back from TEXT
		ID:           uuid.MustParse(m.ID),
		Username:     m.Username,
		PasswordHash: m.PasswordHash,
		CreatedAt:    m.CreatedAt,
	}
}
