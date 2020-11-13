package datasource

const (
	DefaultTimeoutMs        uint64 = 4000
	DefaultListenIntervalMs uint64 = 5000
)

type Config struct {
	TimeoutMs        uint64 `yaml:"timeoutMs"`
	ListenIntervalMs uint64 `yaml:"listenIntervalMs"`
}
