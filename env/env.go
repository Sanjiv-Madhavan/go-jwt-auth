package env

import (
	"context"

	envconfig "github.com/sethvargo/go-envconfig"
)

type EnvironmentSettings struct {
	HTTPServerPort string `env:"HTTP_SERVER_PORT, default=8080"`
	MongoDBURL     string `env:"MONGO_DB_URL, default=mongodb+srv://root:root@cluster0.5x3zq.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0"`
	SECRET_KEY     string `env:"SECRET_KEY"`
}

var Environment EnvironmentSettings

func LoadEnvironment() error {
	return envconfig.Process(context.Background(), &Environment)
}
