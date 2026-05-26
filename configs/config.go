package configs

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	HTTP            HTTPConfig            `yaml:"http"`
	Metrics         MetricsConfig         `yaml:"metrics"`
	Kafka           KafkaConfig           `yaml:"kafka"`
	Business        BusinessConfig        `yaml:"business"`
	AnomalyDetector AnomalyDetectorConfig `yaml:"anomaly_detector"`
	Stoplist        StoplistConfig        `yaml:"stoplist"`
	Logging         LoggingConfig         `yaml:"logging"`
}

type HTTPConfig struct {
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
}

type MetricsConfig struct {
	Port string `yaml:"port"`
	Path string `yaml:"path"`
}

type KafkaConfig struct {
	Brokers []string `yaml:"brokers"`
	Topic   string   `yaml:"topic"`
	GroupID string   `yaml:"group_id"`
}

type BusinessConfig struct {
	WindowSize  time.Duration `yaml:"window_size"`
	BucketSize  time.Duration `yaml:"bucket_size"`
	TopCacheTTL time.Duration `yaml:"top_cache_ttl"`
	TopMaxSize  int           `yaml:"top_max_size"`
}

type AnomalyDetectorConfig struct {
	RequestsPerSecond int    `yaml:"requests_per_second"`
	Burst             int    `yaml:"burst"`
	KeyType           string `yaml:"key_type"`
}

type StoplistConfig struct {
	PersistFile string `yaml:"persist_file"`
}

type LoggingConfig struct {
	Level     string `yaml:"level"`
	Format    string `yaml:"format"`
	AddSource bool   `yaml:"add_source"`
}

func DefaultConfig() *Config {
	return &Config{
		HTTP: HTTPConfig{
			Port:         8080,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		Metrics: MetricsConfig{
			Port: "9090",
			Path: "/metrics",
		},
		Kafka: KafkaConfig{
			Brokers: []string{"localhost:9092"},
			Topic:   "search-events",
			GroupID: "trending-service",
		},
		Business: BusinessConfig{
			WindowSize:  5 * time.Minute,
			BucketSize:  1 * time.Second,
			TopCacheTTL: 1 * time.Second,
			TopMaxSize:  100,
		},
		AnomalyDetector: AnomalyDetectorConfig{
			RequestsPerSecond: 100,
			Burst:             100,
			KeyType:           "user_query",
		},
		Stoplist: StoplistConfig{
			PersistFile: "./data/stoplist.txt",
		},
		Logging: LoggingConfig{
			Level:     "info",
			Format:    "json",
			AddSource: false,
		},
	}
}
func Load(configPath string) (*Config, error) {
	cfg := DefaultConfig()

	if configPath == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.HTTP.Port == 0 {
		return fmt.Errorf("http.port is required")
	}
	if len(c.Kafka.Brokers) == 0 {
		return fmt.Errorf("kafka.brokers is required")
	}
	if c.Kafka.Topic == "" {
		return fmt.Errorf("kafka.topic is required")
	}
	if c.Business.WindowSize < c.Business.BucketSize {
		return fmt.Errorf("window_size (%v) must be >= bucket_size (%v)",
			c.Business.WindowSize, c.Business.BucketSize)
	}
	if c.AnomalyDetector.RequestsPerSecond <= 0 {
		return fmt.Errorf("anomaly_detector.requests_per_second must be > 0")
	}
	return nil
}
