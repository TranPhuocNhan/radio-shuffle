package mq

// Config describes the RabbitMQ topology and client behavior used by the app.
type Config struct {
	URL           string
	Exchange      string
	Queue         string
	RetryQueue    string
	DLQ           string
	RetryTTLMS    int
	MaxRetries    int
	RetryRouteKey string
	MainRouteKey  string
	DLQRouteKey   string
	PrefetchCount int
}
