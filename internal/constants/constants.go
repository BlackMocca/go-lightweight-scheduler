package constants

type AuthAdapter string

const (
	AUTH_ADAPTER_BASIC_AUTH AuthAdapter = "basicauth"
	AUTH_ADAPTER_APIKEY     AuthAdapter = "apikey"
)

type TriggerType string

const (
	TRIGGER_TYPE_SCHEDULE TriggerType = "SCHEDULE"
	TRIGGER_TYPE_EXTERNAL TriggerType = "EXTERNAL"
)
