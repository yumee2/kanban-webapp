package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	HttpServer      `yaml:"http_server"`
	PostgresConnect `yaml:"postgres_storage"`
}

type HttpServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

type PostgresConnect struct {
	Host         string `yaml:"host" env-default:"localhost"`
	Port         int    `yaml:"port" env-default:"5432"`
	User         string `yaml:"user" env-default:"postgres"`
	Password     string `yaml:"password"  env-required:"true"`
	DatabaseName string `yaml:"dbname"  env-required:"true"`
}

func MustLoad() *Config {
	configPath := "config/config.yaml"
	if value := os.Getenv("CONFIG_PATH"); value != "" {
		configPath = value
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", err)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}
