package repositories

import (
	"context"
	"database/sql"

	"github.com/mlvieira/nsfwdetection/internal/models"
)

type UserRepository interface {
	CheckLogin(ctx context.Context, username, password string) (models.User, error)
	AddUser(ctx context.Context, u models.User) error
}

type UploadedRepository interface {
	ListUploadsCursor(ctx context.Context, cursorID, limit int, reviewed *bool) ([]models.UploadedImage, error)
	LabelUpload(ctx context.Context, hash, label string) (int, error)
	UploadImage(ctx context.Context, img models.UploadedImage) error
	ListTotalUploads(ctx context.Context, reviewed *bool) (int, error)
	GetFilePathByHash(ctx context.Context, hash string) (string, error)
	DeleteImage(ctx context.Context, hash string) (int, error)
}

type Repositories struct {
	User     UserRepository
	Uploaded UploadedRepository
}

func NewRepositories(conn *sql.DB) *Repositories {
	return &Repositories{
		User:     NewUserRepository(conn),
		Uploaded: NewUploadedRepository(conn),
	}
}
