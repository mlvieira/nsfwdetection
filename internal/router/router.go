package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/mlvieira/nsfwdetection/internal/config"
	"github.com/mlvieira/nsfwdetection/internal/driver/redis"
	"github.com/mlvieira/nsfwdetection/internal/handlers"
	"github.com/mlvieira/nsfwdetection/internal/middleware"
	"github.com/mlvieira/nsfwdetection/internal/repositories"
	"github.com/mlvieira/nsfwdetection/internal/services"
	"github.com/mlvieira/nsfwdetection/internal/websockets"
)

func SetupRoutes(repositories *repositories.Repositories, redisClient *redis.RedisClient) http.Handler {
	mux := chi.NewRouter()

	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{config.AppConfig.Server.DomainName},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Forwarded-For"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	hub := websockets.NewHub()
	go hub.Run()

	nsfwService := services.NewNSFWService(redisClient, hub, repositories)
	apiService := services.NewAPIService(hub, repositories)
	handlersInstance := handlers.NewHandlers(repositories, hub)
	nsfwHandlers := handlers.NewNSFWHandlers(handlersInstance, nsfwService)
	apiHandlers := handlers.NewAPIHandlers(handlersInstance, apiService)

	mux.Get("/ws", handlers.HandleWebSocket(hub))

	mux.Route("/api", func(r chi.Router) {
		r.Post("/detect-nsfw", nsfwHandlers.NSFWHandler)
	})

	mux.Options("/*", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.Route("/admin", func(r chi.Router) {
		r.Post("/login", apiHandlers.Login)

		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuth)

			r.Get("/images", apiHandlers.PaginationUploads)
			r.Post("/label/add/{hash}", apiHandlers.LabelImage)
			r.Post("/label/update/{hash}", apiHandlers.LabelImage)
			r.Post("/delete/{hash}", apiHandlers.DeleteImage)
		})
	})

	return mux
}
