package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/fhsmendes/rate-limiter/configs"
	"github.com/fhsmendes/rate-limiter/infra/webserver/handler"
	"github.com/go-chi/chi"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Printf("Erro ao carregar o arquivo .env: %v\n", err)
		return
	}

	configs.LoadEnvs()

	authHandler := handler.NewAuthHandler()

	r := chi.NewRouter()
	r.Post("/token", authHandler.GetJWT)

	fmt.Printf("aplicação iniciada na porta %d", configs.Envs.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", configs.Envs.Port), r); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
}
