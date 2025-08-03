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
	Ip_Block_Duration_Seconds      int    `env:"IP_BLOCK_DURATION_SECONDS" envDefault:"300"`
	Jwt_Secret                     string `env:"JWT_SECRET" envDefault:""`
	Jwt_Expires_In                 int    `env:"JWT_EXPIRES_IN" envDefault:"300"`
	Redis_Enabled                  bool   `env:"REDIS_ENABLED" envDefault:"false"`
	Redis_Host                     string `env:"REDIS_HOST" envDefault:"localhost:6379"`
	Redis_Password                 string `env:"REDIS_PASSWORD" envDefault:""`
	Redis_DB                       int    `env:"REDIS_DB" envDefault:"0"`

	Token_Auth *jwtauth.JWTAuth
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
