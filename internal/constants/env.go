package constants

import "os"

func getEnv(key string, defaultValue string) string {
	val, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return val
}

var (
	ENV_APP_PORT         = getEnv("APP_PORT", "3000")
	ENV_DATABASE_ADAPTER = getEnv("DATABASE_ADAPTER", "postgres")
	ENV_DATABASE_URL     = getEnv("DATABASE_URL", "")
)
