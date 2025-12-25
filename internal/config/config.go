package config

import "os"

type Config struct {
	HTTPAddr    string
	DatabaseURL string
	JWTSecret   string
}

func FromEnv() Config {
	return Config{
		HTTPAddr:    getenv("HTTP_ADDR", ":8080"),
		DatabaseURL: must("DATABASE_URL"),
		JWTSecret:   must("JWT_SECRET"),
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func must(k string) string {
	v := os.Getenv(k)
	if v == "" {
		panic("missing env var: " + k)
	}
	return v
}
