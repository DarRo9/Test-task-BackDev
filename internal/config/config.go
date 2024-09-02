package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

type JWT struct {
	AccessTokenTTL  time.Duration `yaml:"access_token_ttl"`
	RefreshTokenTTL time.Duration `yaml:"refresh_token_ttl"`
	SigningKey      string
}

type Mongo struct {
	URI      string
	User     string
	Password string
	Name     string
	Database string
}

type Config struct {
	Env        string `yaml:"env" env-default:"local"`
	HTTPServer `yaml:"http_server"`
	Mongo
	JWT `yaml:"jwt"`
}

func MakeEnvSettings(config *Config) {
	if config.Env == "local" {
		config.Mongo.URI = "mongodb://localhost:27017"
	} else {
		config.Mongo.URI = os.Getenv("MONGO_URI")
	}

	config.Mongo.Database = os.Getenv("MONGO_DATABASE")

	config.JWT.SigningKey = os.Getenv("JWT_SIGNING_KEY")
}

func Loading() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var config Config

	if err := cleanenv.ReadConfig(configPath, &config); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	MakeEnvSettings(&config)

	return &config
}
