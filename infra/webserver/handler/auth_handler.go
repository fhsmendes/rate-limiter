package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/fhsmendes/rate-limiter/configs"
	"github.com/go-chi/jwtauth"
)

type GetJwtIpunt struct {
	RateLimit         int `json:"rate_limit"`
	RateLimitDuration int `json:"rate_limit_duration"`
}

type AuthHandler struct {
	Jwt *jwtauth.JWTAuth
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		Jwt: configs.Envs.Token_Auth, // Aqui que você garante que não é nil
	}
}

func (a *AuthHandler) GetJWT(w http.ResponseWriter, r *http.Request) {
	var inputData GetJwtIpunt
	err := json.NewDecoder(r.Body).Decode(&inputData)
	if err != nil {
		http.Error(w, "Invalid input data", http.StatusBadRequest)
		return
	}

	_, tokenString, _ := a.Jwt.Encode(map[string]interface{}{
		"rate_limit":          inputData.RateLimit,
		"rate_limit_duration": inputData.RateLimitDuration,
		"exp":                 time.Now().Add(time.Duration(configs.Envs.Jwt_Expires_In) * time.Second).Unix(),
	})

	apiKey := struct {
		API_KEY string `json:"api_key"`
	}{
		API_KEY: tokenString,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(apiKey)

}
