package transport

type Config struct {
	// TimeoutMs is the maximum amount of time a client will wait for a connect to complete
	TimeoutMs uint64 `yaml:"timeout"`
	// Secure is setting the socket encrypted or not
	Secure bool
}
