package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env                 string              `env:"APP_ENV" env-default:"local"`
	HTTPServer          HTTPServerConfig    `env-prefix:"SERVER_"`
	DB                  DBConfig            `env-prefix:"DB_"`
	Redis               RedisConfig         `env-prefix:"REDIS_"`
	YandexTranslatorAPI TranslatorAPIConfig `env-prefix:"TRANSLATOR_API_"`
}

type HTTPServerConfig struct {
	Address     string        `env:"ADDRESS" env-required:"true"`
	Timeout     time.Duration `env:"TIMEOUT" env-default:"4s"`
	IdleTimeout time.Duration `env:"IDLE_TIMEOUT" env-default:"60s"`
}

type DBConfig struct {
	Host     string `env:"HOST" env-default:"localhost"`
	Port     string `env:"PORT" env-default:"5432"`
	User     string `env:"USER" env-required:"true"`
	Password string `env:"PASSWORD" env-required:"true"`
	Name     string `env:"NAME" env-required:"true"`
}

type RedisConfig struct {
	Host     string `env:"HOST" env-default:"localhost"`
	Port     string `env:"PORT" env-default:"6379"`
	Password string `env:"PASSWORD" env-required:"true"`
}

type TranslatorAPIConfig struct {
	Key string `env:"KEY" env-required:"true"`
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
