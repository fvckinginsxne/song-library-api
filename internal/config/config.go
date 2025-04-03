package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env                 string              `yaml:"env" env-default:"local"`
	HTTPServer          HTTPServer          `yaml:"http_server"`
	DB                  DB                  `yaml:"db"`
	DeepSeekAPI         DeepSeekAPI         `yaml:"deepseek_api"`
	YandexTranslatorAPI YandexTranslatorAPI `yaml:"yandex_translator_api"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-required:"true"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

type DB struct {
	Host     string `yaml:"host" env-required:"true"`
	Port     string `yaml:"port" env-required:"true"`
	Username string `yaml:"username" env-required:"true"`
	Password string `yaml:"password" env-required:"true"`
	Name     string `yaml:"name" env-required:"true"`
}

type DeepSeekAPI struct {
	Key string `yaml:"key" env-required:"true"`
}

type YandexTranslatorAPI struct {
	Key string `yaml:"key" env-required:"true"`
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
