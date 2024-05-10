package env

import "github.com/joho/godotenv"

func LoadEnvs() {
	// Load .env file
	godotenv.Load(".env")
}
