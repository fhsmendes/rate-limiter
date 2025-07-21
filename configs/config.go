package configs

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/go-chi/jwtauth"
)

type envs struct {
	Port                           int    `env:"PORT" envDefault:"8080"`
	Ip_Rate_Limit                  int    `env:"IP_RATE_LIMIT" envDefault:"10"`
	Ip_Rate_Limit_Duration_Seconds int    `env:"IP_RATE_LIMIT_DURATION_SECONDS" envDefault:"60"`
	Jwt_Secret                     string `env:"JWT_SECRET" envDefault:""`
	Jwt_Expires_In                 int    `env:"JWT_EXPIRES_IN" envDefault:"300"`
	Token_Auth                     *jwtauth.JWTAuth
}

var (
	Envs envs
)

func LoadEnvs() {
	if err := env.Parse(&Envs); err != nil {
		panic(fmt.Errorf("erro ao carregar as variáveis de ambiente: %v", err))
	}

	Envs.Token_Auth = jwtauth.New("HS256", []byte(Envs.Jwt_Secret), nil)
}
