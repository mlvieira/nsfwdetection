package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mlvieira/nsfwdetection/internal/config"
	"github.com/mlvieira/nsfwdetection/internal/driver/mysql"
	"github.com/mlvieira/nsfwdetection/internal/driver/redis"
	"github.com/mlvieira/nsfwdetection/internal/logger"
	"github.com/mlvieira/nsfwdetection/internal/repositories"
	"github.com/mlvieira/nsfwdetection/internal/router"
	"github.com/mlvieira/nsfwdetection/internal/tfmodel"
	"github.com/mlvieira/nsfwdetection/internal/worker"
)

func main() {
	config.LoadConfig("./config.toml")

	if err := logger.Init("logs/app.log"); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	if err := tfmodel.LoadModel(config.AppConfig.Model.ModelPath); err != nil {
		logger.Fatalf("Failed to load model: %v", err)
	}
	defer tfmodel.SharedNSFWModel.Close()

	worker.InitWorkerPool(tfmodel.SharedNSFWModel)
	defer worker.ShutdownWorkerPool()

	conn, err := mysql.OpenDB()
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close()

	redisClient := redis.NewRedisClient(
		config.AppConfig.Redis.Addr,
		config.AppConfig.Redis.Password,
		config.AppConfig.Redis.DB,
	)

	repositories := repositories.NewRepositories(conn)

	mux := router.SetupRoutes(repositories, redisClient)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.AppConfig.Server.Port),
		Handler: mux,
	}

	logger.Info("Starting server on port %d", config.AppConfig.Server.Port)
	if err = server.ListenAndServe(); err != nil {
		logger.Fatalf("Server failed: %v", err)
	}
}
