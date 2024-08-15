package fiatramps

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

func GetEnvConfig(key string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		log.Fatalf("%s not set\n", key)
	}

	return val
}

func LoadEnv(envFile string) {
	err := godotenv.Load(envFile)

	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}
