package constants

import (
	"os"

	"github.com/spf13/cast"
)

func getEnv(key string, defaultValue string) string {
	val, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return val
}

var (
	ENV_APP_PORT                     = getEnv("APP_PORT", "3000")
	ENV_API_AUTH_ADAPTER             = getEnv("API_AUTH_ADAPTER", "basicauth")
	ENV_API_AUTH_BASIC_AUTH_USERNAME = getEnv("API_AUTH_BASIC_AUTH_USERNAME", "")
	ENV_API_AUTH_BASIC_AUTH_PASSWORD = getEnv("API_AUTH_BASIC_AUTH_PASSWORD", "")
	ENV_API_AUTH_API_KEY_NAME        = getEnv("API_AUTH_API_KEY_NAME", "")
	ENV_API_AUTH_API_KEY_VALUE       = getEnv("API_AUTH_API_KEY_VALUE", "")
	ENV_MIGRATE_TABLE                = cast.ToBool(getEnv("MIGRATE_TABLE", "true"))
	ENV_DATABASE_ADAPTER             = getEnv("DATABASE_ADAPTER", "postgres")
	ENV_DATABASE_URL                 = getEnv("DATABASE_URL", "")
	ENV_ENABLED_DAG_EXAMPLE          = cast.ToBool(getEnv("ENABLED_DAG_EXAMPLE", "true"))
)
