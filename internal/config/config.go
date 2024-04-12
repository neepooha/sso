package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env      string        `yaml:"env" env-default:"local"`
	TokenTTL time.Duration `yaml:"token_ttl" env-default:"1h"`
	GRPC     `yaml:"grpc"`
	Storage  `yaml:"storage"`
}
type Storage struct {
	Host            string `yaml:"host" env-required:"true"`
	Port            string `yaml:"port" env-required:"true"`
	User            string `yaml:"user" env-required:"true"`
	Dbname          string `yaml:"dbname" env-required:"true"`
	Password        string `yaml:"password" env-required:"true"`
	Migrations_path string `yaml:"migrations_path" env-required:"true"`
}

type GRPC struct {
	Host    string        `yaml:"host" env-default:""`
	Port    string        `yaml:"port" env-default:"44044"`
	Timeout time.Duration `yaml:"timeout" env-default:"15s"`
}

// Must - обозначает, что функция либо выполнится, либо вызовет панику
func MustLoad() *Config {
	// loads environment variables from the .env file
	if err := godotenv.Load("./config.env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	// get configPath from our new env
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}
	return MustLoadByPath(configPath)
}

func MustLoadByPath(configPath string) *Config {
	// check if the file exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatal("config file doesn't exist: ", configPath)
	}

	// read config from yaml
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatal("can't read config ", err)
	}

	return &cfg
}
