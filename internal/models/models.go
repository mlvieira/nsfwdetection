package models

import "time"

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UploadedImage struct {
	ID         int       `json:"id"`
	FilePath   string    `json:"filepath"`
	FileHash   string    `json:"filehash"`
	Label      string    `json:"label"`
	NewLabel   string    `json:"new_label"`
	Confidence float32   `json:"confidence"`
	Reviewed   bool      `json:"reviewed"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
