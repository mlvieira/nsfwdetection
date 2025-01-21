package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/mlvieira/nsfwdetection/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type userRepo struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepo{db: db}
}

func (ur *userRepo) AddUser(ctx context.Context, u models.User) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	txn, err := ur.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			txn.Rollback()
		}
	}()

	query := `INSERT INTO users
			(username, password, created_at, updated_at)
			VALUES
			(?, ?, ?, ?)
	`
	_, err = txn.Exec(query,
		u.Username,
		u.Password,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	if err = txn.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (ur *userRepo) CheckLogin(ctx context.Context, username, password string) (models.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var user models.User
	var hashedPassword string

	query := `
		SELECT id, username, password FROM users WHERE username = ?
	`

	err := ur.db.QueryRowContext(ctx, query, username).Scan(&user.ID, &user.Username, &hashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return user, fmt.Errorf("invalid username or password")
		}
		return user, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return user, fmt.Errorf("invalid username or password")
	}

	return user, nil
}
