package configs

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Error loading .env file")
	}
}

func Get(key string) string {
	return os.Getenv(key)
}
