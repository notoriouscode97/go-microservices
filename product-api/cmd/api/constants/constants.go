package constants

var EnvKeys = envKeys{
	Env:               "ENV",
	ServerAddress:     "SERVER_ADDRESS",
	DatabaseDSN:       "DATABASE_DSN",
	CorsAllowedOrigin: "CORS_ALLOWED_ORIGIN",
}

type envKeys struct {
	Env               string
	ServerAddress     string
	DatabaseDSN       string
	CorsAllowedOrigin string
}
