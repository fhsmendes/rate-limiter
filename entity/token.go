package entity

import (
	"github.com/fhsmendes/rate-limiter/configs"
	"github.com/go-chi/jwtauth"
)

type TokenAuth struct {
	RateLimit         int `json:"rate_limit"`
	RateLimitDuration int `json:"rate_limit_duration"`
	BlockDuration     int `json:"block_duration"`
}

func (a *TokenAuth) Decode(tokenString string) error {
	jwt := jwtauth.New("HS256", []byte(configs.Envs.Jwt_Secret), nil)
	token, err := jwt.Decode(tokenString)
	if err != nil {
		return err
	}

	claims := token.PrivateClaims()

	rateLimit, _ := claims["rate_limit"].(float64)
	rateLimitDuration, _ := claims["rate_limit_duration"].(float64)
	blockDuration, _ := claims["block_duration"].(float64)

	a.RateLimit = int(rateLimit)
	a.RateLimitDuration = int(rateLimitDuration)
	a.BlockDuration = int(blockDuration)

	return nil
}
