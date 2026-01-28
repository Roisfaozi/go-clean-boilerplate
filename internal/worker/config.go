package worker

// WorkerConfig holds configuration specifically for the worker package
// to avoid import cycles with internal/config
type WorkerConfig struct {
	SMTP SMTPConfig
}

type SMTPConfig struct {
	Host       string
	Port       int
	Username   string
	Password   string
	FromSender string
	FromEmail  string
}
