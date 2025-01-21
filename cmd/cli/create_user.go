package main

import (
	"context"
	"fmt"
	"os"

	"github.com/mlvieira/nsfwdetection/internal/config"
	"github.com/mlvieira/nsfwdetection/internal/driver/mysql"
	"github.com/mlvieira/nsfwdetection/internal/logger"
	"github.com/mlvieira/nsfwdetection/internal/models"
	"github.com/mlvieira/nsfwdetection/internal/repositories"
	"github.com/mlvieira/nsfwdetection/internal/utils"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: create_user <username> <password>")
		os.Exit(1)
	}

	config.LoadConfig("./config.toml")

	conn, err := mysql.OpenDB()
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close()

	repositories := repositories.NewRepositories(conn)

	username := os.Args[1]
	password := os.Args[2]

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		fmt.Println("Error hashing password:", err)
		os.Exit(1)
	}

	user := models.User{
		Username: username,
		Password: hashedPassword,
	}

	ctx := context.Background()
	if err := repositories.User.AddUser(ctx, user); err != nil {
		fmt.Println("Failed to create user:", err)
		os.Exit(1)
	}

	fmt.Println("User created successfully!")
}
