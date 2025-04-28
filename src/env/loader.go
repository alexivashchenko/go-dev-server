package env

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func Load() {

	if _, err := os.Stat(".env"); err == nil {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

}
