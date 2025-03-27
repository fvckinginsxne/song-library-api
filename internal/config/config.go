package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string `env:"ENV" env-default:"local"`
	HTTPServer HTTPServer
	DB         DB
	GeniusAPI  GeniusAPI
}

type HTTPServer struct {
	Address     string        `env:"SERVER_ADDRESS" env-required:"true"`
	Timeout     time.Duration `env:"SERVER_TIMEOUT" env-default:"4s"`
	IdleTimeout time.Duration `env:"SERVER_IDLE_TIMEOUT" env-default:"60s"`
}

type DB struct {
	Host     string `env:"DB_HOST" env-required:"true"`
	Port     string `env:"DB_PORT" env-required:"true"`
	Username string `env:"DB_USERNAME" env-required:"true"`
	Password string `env:"DB_PASSWORD" env-required:"true"`
	Name     string `env:"DB_NAME" env-required:"true"`
}

type GeniusAPI struct {
	AccessToken string `env:"GENIUS_ACCESS_TOKEN" env-required:"true"`
	BaseURL     string `env:"GENIUS_BASE_URL" env-required:"true"`
}

// MustLoad Load config file and panic if errors occurs
func MustLoad() *Config {
	path := fetchConfigPath()
	if path == "" {
		panic("config file path is empty")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file not found: " + path)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}

	return &cfg
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	if err := flag.CommandLine.Parse(os.Args[1:2]); err != nil {
		panic(err)
	}

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
