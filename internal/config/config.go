package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env         string `yaml:"env" env-default:"local"`
	DatabaseURL string `yaml:"database_url" env:"DATABASE_URL"`
	HTTPServer  `yaml:"http_server"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
	User        string        `yaml:"user" env-required:"true"`
	Password    string        `yaml:"password" env-required:"true" env:"HTTP_SERVER_PASSWORD"`
}

// MustLoad загружает конфигурацию из файла или завершает программу с ошибкой
// Возвращает указатель на загруженную конфигурацию
func MustLoad() *Config {

	err := godotenv.Load()
	fmt.Println(err)
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) { // Проверка существования файла
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	// Читаем конфигурацию из файла с помощью cleanenv
	// cleanenv.ReadConfig парсит YAML файл и заполняет структуру Config
	// Если произошла ошибка - завершаем программу
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		cfg.DatabaseURL = dbURL
		log.Println("Using DATABASE_URL from environment")
	}

	// Явная загрузка из переменных окружения (может перезаписать значения из YAML)
	if user := os.Getenv("AUTH_USER"); user != "" {
		cfg.User = user
		log.Println("Using AUTH_USER from environment")
	}

	if password := os.Getenv("AUTH_PASSWORD"); password != "" {
		cfg.Password = password
		log.Println("Using AUTH_PASSWORD from environment")
	}

	if cfg.User == "" || cfg.Password == "" {
		log.Fatal("Username and Password is empty")
	}

	return &cfg
}
