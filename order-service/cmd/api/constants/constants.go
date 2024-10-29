package constants

var EnvKeys = envKeys{
	Env:               "ENV",
	ServerAddress:     "SERVER_ADDRESS",
	DatabaseDSN:       "DATABASE_DSN",
	CorsAllowedOrigin: "CORS_ALLOWED_ORIGIN",
	SmtpHost:          "SMTP_HOST",
	SmtpPort:          "SMTP_PORT",
	SmtpUsername:      "SMTP_USERNAME",
	SmtpPassword:      "SMTP_PASSWORD",
	SmtpSender:        "SMTP_SENDER",
}

type envKeys struct {
	Env               string
	ServerAddress     string
	DatabaseDSN       string
	CorsAllowedOrigin string
	SmtpHost          string
	SmtpPort          string
	SmtpUsername      string
	SmtpPassword      string
	SmtpSender        string
}
