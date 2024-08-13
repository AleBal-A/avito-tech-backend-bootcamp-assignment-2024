package config

import (
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Logger   LoggerConfig   `yaml:"logger"`
	Auth     AuthConfig     `yaml:"auth"`
}

type ServerConfig struct {
	Port        string        `yaml:"port"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
	//User        string
	//Password    string
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type LoggerConfig struct {
	Level string `yaml:"level"`
}

type AuthConfig struct {
	JWTSecret string `yaml:"jwt_secret"`
}

func MustLoad() *Config {
	path := fetchConfigPath()
	if path == "" {
		panic("config path is empty. \n" +
			"Set the CONFIG_PATH environment variable or use the `--config=\"path/to/config.yaml\"` flag.")
	}
	return MustLoadByPath(path)
}

// fetchConfigPath fetches config path from command line flag or environment variable.
// Priority: flag > environment > default.
// Default value is empty string
func fetchConfigPath() string {
	var res string

	// EX: --config="path/to/conf.yaml"
	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()
	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}

func MustLoadByPath(configPath string) *Config {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("cannot read config: " + err.Error())
	}

	// JWT load
	if cfg.Auth.JWTSecret == "" {
		cfg.Auth.JWTSecret = os.Getenv("JWT_SECRET")
		if cfg.Auth.JWTSecret == "" {
			panic("JWT_SECRET is not set in config or ENV")
		} else {
			log.Println("JWT_SECRET loaded from ENV")
		}
	} else {
		log.Println("JWT_SECRET loaded from config")
	}

	// DB load
	if cfg.Database.Password == "" {
		cfg.Database.Password = os.Getenv("DB_PASSWORD")
		if cfg.Database.Password == "" {
			panic("DB_PASSWORD is not set in config or ENV")
		} else {
			log.Println("DB_PASSWORD loaded from ENV")
		}
	} else {
		log.Println("DB_PASSWORD loaded from config")
	}

	return &cfg
}
