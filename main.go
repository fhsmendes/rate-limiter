package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/fhsmendes/rate-limiter/configs"
	"github.com/fhsmendes/rate-limiter/database"
	"github.com/fhsmendes/rate-limiter/limiter"
	"github.com/fhsmendes/rate-limiter/webserver/handler"

	mid "github.com/fhsmendes/rate-limiter/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Printf("error loading .env file: %v\n", err)
		return
	}

	configs.LoadEnvs()

	var rateLimiter limiter.RateLimiter

	if configs.Envs.Redis_Enabled {
		redisClient, err := database.NewRedisClient()
		if err != nil {
			log.Fatalf("tried to connect to Redis: %v", err)
		}

		fmt.Println("using redis")
		rateLimiter = database.NewRedisLimiter(redisClient)
	} else {
		fmt.Println("using memory")
		rateLimiter = database.NewMemoryLimiter()
	}

	authHandler := handler.NewAuthHandler()
	testHandler := handler.NewTestHandler()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.AllowContentType("application/json"))
	r.Use(middleware.NoCache)

	r.Use(mid.RateLimitMiddleware(rateLimiter))

	r.Post("/token", authHandler.GetJWT)

	r.Route("/test", func(r chi.Router) {
		r.Use(mid.ValidTokenMiddleware)
		r.Get("/", testHandler.GetTest)
	})

	fmt.Printf("start on %d\n", configs.Envs.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", configs.Envs.Port), r); err != nil {
		log.Fatalf("error on start server: %v", err)
	}
}
