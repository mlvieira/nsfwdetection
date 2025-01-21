package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/mlvieira/nsfwdetection/internal/logger"
	"github.com/mlvieira/nsfwdetection/internal/models"
)

type uploadedRepo struct {
	db *sql.DB
}

func NewUploadedRepository(db *sql.DB) UploadedRepository {
	return &uploadedRepo{db: db}
}

func (u *uploadedRepo) UploadImage(ctx context.Context, img models.UploadedImage) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	txn, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			txn.Rollback()
		}
	}()

	query := `INSERT INTO uploaded_images
			(file_path, file_hash, label, confidence, reviewed, created_at, updated_at)
			VALUES
			(?, ?, ?, ?, ?, ?, ?)
	`
	_, err = txn.Exec(query,
		img.FilePath,
		img.FileHash,
		img.Label,
		img.Confidence,
		img.Reviewed,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		if sqlErr, ok := err.(*mysql.MySQLError); ok && sqlErr.Number == 1062 {
			logger.Info("File hash already exists, skipping insertion: %s", img.FileHash)
			return nil
		} else {
			return fmt.Errorf("failed to insert uploaded image: %w", err)
		}
	}

	if err = txn.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (u *uploadedRepo) ListUploadsCursor(ctx context.Context, cursorID, limit int, reviewed *bool) ([]models.UploadedImage, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `
		SELECT 
			id, file_path, file_hash, label, new_label, confidence, reviewed, created_at 
		FROM 
			uploaded_images
		WHERE 
			id > ?
	`

	var args []interface{}
	args = append(args, cursorID)

	if reviewed != nil {
		query += " AND reviewed = ?"
		args = append(args, *reviewed)
	}

	query += `
		ORDER BY
			id DESC
		LIMIT ?
	`
	args = append(args, limit)

	rows, err := u.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var uploads []models.UploadedImage
	for rows.Next() {
		var upload models.UploadedImage
		err := rows.Scan(
			&upload.ID,
			&upload.FilePath,
			&upload.FileHash,
			&upload.Label,
			&upload.NewLabel,
			&upload.Confidence,
			&upload.Reviewed,
			&upload.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		uploads = append(uploads, upload)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return uploads, nil
}

// LabelUpload updates the label for an image
func (u *uploadedRepo) LabelUpload(ctx context.Context, hash, label string) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	txn, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			txn.Rollback()
		}
	}()

	query := `UPDATE uploaded_images SET new_label = ?, updated_at = ?, reviewed = true WHERE file_hash = ?`
	result, err := txn.Exec(query, label, time.Now(), hash)
	if err != nil {
		return 0, fmt.Errorf("failed to execute query: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to fetch affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return int(rowsAffected), fmt.Errorf("no rows updated, image with hash %s not found", hash)
	}

	if err = txn.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return int(rowsAffected), nil
}

func (u *uploadedRepo) ListTotalUploads(ctx context.Context, reviewed *bool) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var count int

	query := `
		SELECT 
			count(1) 
		FROM 
			uploaded_images
	`

	var args []interface{}
	if reviewed != nil {
		query += " WHERE reviewed = ?"
		args = append(args, *reviewed)
	}

	if err := u.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (u *uploadedRepo) GetFilePathByHash(ctx context.Context, hash string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	var filePath string

	query := `
		SELECT 
			file_path 
		FROM 
			uploaded_images
		WHERE 
			file_hash = ?
		LIMIT 1
	`

	if err := u.db.QueryRowContext(ctx, query, hash).Scan(&filePath); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}

	return filePath, nil
}

func (u *uploadedRepo) DeleteImage(ctx context.Context, hash string) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	txn, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			txn.Rollback()
		}
	}()

	query := `DELETE FROM uploaded_images WHERE file_hash = ?`
	result, err := txn.Exec(query, hash)
	if err != nil {
		return 0, fmt.Errorf("failed to delete hash: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if rowsAffected == 0 {
		return 0, fmt.Errorf("no rows deleted, image with hash %s not found", hash)
	}

	if err = txn.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return int(rowsAffected), nil
}
