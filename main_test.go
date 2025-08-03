package main

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/fhsmendes/rate-limiter/configs"
	"github.com/fhsmendes/rate-limiter/database"
	"github.com/fhsmendes/rate-limiter/entity"
	mid "github.com/fhsmendes/rate-limiter/middleware"
	"github.com/go-chi/jwtauth"
	"github.com/go-redis/redis/v8"
)

func TestRateLimiterWithRedis(t *testing.T) {
	os.Setenv("REDIS_ENABLED", "true")
	os.Setenv("IP_RATE_LIMIT", "2")
	os.Setenv("IP_RATE_LIMIT_DURATION_SECONDS", "5")
	os.Setenv("IP_BLOCK_DURATION_SECONDS", "10")
	os.Setenv("JWT_SECRET", "SECRET")
	os.Setenv("JWT_EXPIRES_IN", "120")
	defer os.Unsetenv("REDIS_ENABLED")
	defer os.Unsetenv("IP_RATE_LIMIT")
	defer os.Unsetenv("IP_RATE_LIMIT_DURATION_SECONDS")
	defer os.Unsetenv("IP_BLOCK_DURATION_SECONDS")
	defer os.Unsetenv("JWT_SECRET")
	defer os.Unsetenv("JWT_EXPIRES_IN")

	configs.LoadEnvs()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		t.Fatalf("não foi possível conectar ao Redis: %v. Verifique se o servidor Redis está em execução.", err)
	}
	defer rdb.Close()
	limiter := database.NewRedisLimiter(rdb)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handlerWithMiddleware := mid.RateLimitMiddleware(limiter)(testHandler)

	t.Run("Blocked by limit (IP)", func(t *testing.T) {
		// Limpa o banco de dados do Redis antes de cada sub-teste
		rdb.FlushDB(context.Background())

		limit := 2
		ip := "192.168.1.1"

		// Enviar requisições dentro do limite
		for i := 0; i < limit; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = ip + ":12345"
			rr := httptest.NewRecorder()

			handlerWithMiddleware.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusOK {
				t.Fatalf("Requisição %d foi bloqueada inesperadamente. Status: %v", i+1, status)
			}
		}

		// Enviar uma requisição que excede o limite
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = ip + ":12345"
		rr := httptest.NewRecorder()

		handlerWithMiddleware.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusTooManyRequests {
			t.Fatalf("Requisição excedente não foi bloqueada. Status: %v", status)
		}

		body, _ := ioutil.ReadAll(rr.Body)
		if strings.ReplaceAll(string(body), "\n", "") != strings.ReplaceAll(mid.MsgStatus429, "\n", "") {
			t.Errorf("Mensagem de erro inesperada: %v", string(body))
		}
	})

	t.Run("Blocked by invalid Token", func(t *testing.T) {
		// Limpa o banco de dados do Redis antes de cada sub-teste
		rdb.FlushDB(context.Background())

		token := "test-token-123"

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("API_KEY", token)
		rr := httptest.NewRecorder()

		handlerWithMiddleware.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusUnauthorized {
			t.Fatalf("Requisição inválida não foi bloqueada. Status: %v", status)
		}

		body, _ := ioutil.ReadAll(rr.Body)
		if strings.ReplaceAll(string(body), "\n", "") != strings.ReplaceAll("invalid API key", "\n", "") {
			t.Errorf("Mensagem de erro inesperada: %v", string(body))
		}
	})

	t.Run("Blocked by limit (token)", func(t *testing.T) {
		// Limpa o banco de dados do Redis antes de cada sub-teste
		rdb.FlushDB(context.Background())

		limit := 5
		// Criar token através da API
		inputData := entity.TokenAuth{
			RateLimit:         limit,
			RateLimitDuration: 10,
			BlockDuration:     30,
		}

		jwt := jwtauth.New("HS256", []byte(configs.Envs.Jwt_Secret), nil)
		_, token, _ := jwt.Encode(map[string]interface{}{
			"rate_limit":          inputData.RateLimit,
			"rate_limit_duration": inputData.RateLimitDuration,
			"block_duration":      inputData.BlockDuration,
			"exp":                 time.Now().Add(time.Duration(configs.Envs.Jwt_Expires_In) * time.Second).Unix(),
		})

		// Enviar requisições dentro do limite
		for i := 0; i < limit; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("API_KEY", token)
			rr := httptest.NewRecorder()

			handlerWithMiddleware.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusOK {
				t.Fatalf("Requisição %d foi bloqueada inesperadamente. Status: %v", i+1, status)
			}
		}

		// Enviar uma requisição que excede o limite
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("API_KEY", token)
		rr := httptest.NewRecorder()

		handlerWithMiddleware.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusTooManyRequests {
			t.Fatalf("Requisição inválida não foi bloqueada. Status: %v", status)
		}

		body, _ := ioutil.ReadAll(rr.Body)
		if strings.ReplaceAll(string(body), "\n", "") != strings.ReplaceAll(mid.MsgStatus429, "\n", "") {
			t.Errorf("Mensagem de erro inesperada: %v", string(body))
		}
	})

	t.Run("released after block (token)", func(t *testing.T) {
		// Limpa o banco de dados do Redis antes de cada sub-teste
		rdb.FlushDB(context.Background())

		limit := 5
		// Criar token através da API
		inputData := entity.TokenAuth{
			RateLimit:         limit,
			RateLimitDuration: 10,
			BlockDuration:     15,
		}

		jwt := jwtauth.New("HS256", []byte(configs.Envs.Jwt_Secret), nil)
		_, token, _ := jwt.Encode(map[string]interface{}{
			"rate_limit":          inputData.RateLimit,
			"rate_limit_duration": inputData.RateLimitDuration,
			"block_duration":      inputData.BlockDuration,
			"exp":                 time.Now().Add(time.Duration(configs.Envs.Jwt_Expires_In) * time.Second).Unix(),
		})

		// Enviar requisições dentro do limite
		for i := 0; i < limit; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("API_KEY", token)
			rr := httptest.NewRecorder()

			handlerWithMiddleware.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusOK {
				t.Fatalf("Requisição %d foi bloqueada inesperadamente. Status: %v", i+1, status)
			}
		}

		// Enviar uma requisição que excede o limite
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("API_KEY", token)
		rr := httptest.NewRecorder()

		handlerWithMiddleware.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusTooManyRequests {
			t.Fatalf("Requisição inválida não foi bloqueada. Status: %v", status)
		}

		if status := rr.Code; status == http.StatusTooManyRequests {
			time.Sleep(time.Duration(inputData.BlockDuration+5) * time.Second)
		}

		// Enviar uma requisição após o bloqueio
		req = httptest.NewRequest("GET", "/", nil)
		req.Header.Set("API_KEY", token)
		rr = httptest.NewRecorder()

		handlerWithMiddleware.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Fatalf("Requisição após bloqueio não foi liberada. Status: %v", status)
		}
	})
}
