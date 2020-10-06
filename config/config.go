package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/kelseyhightower/envconfig"
)

const (
	modeDevelopment = "development"
	modeProduction  = "production"
)

var (
	errEndpointRequired            = errors.New("proxy url is required")
	errDatabaseRequired            = errors.New("database credentials are required")
	errIndexWorkerIntervalRequired = errors.New("index worker interval is required")
)

// Config holds the configuration data
type Config struct {
	AppEnv                       string `json:"app_env" envconfig:"APP_ENV" default:"development"`
	ProxyUrl                     string `json:"proxy_url" envconfig:"PROXY_URL"`
	ServerAddr                   string `json:"server_addr" envconfig:"SERVER_ADDR" default:"0.0.0.0"`
	ServerPort                   int64  `json:"server_port" envconfig:"SERVER_PORT" default:"8081"`
	FirstBlockHeight             int64  `json:"first_block_height" envconfig:"FIRST_BLOCK_HEIGHT" default:"1"`
	IndexWorkerInterval          string `json:"index_worker_interval" envconfig:"INDEX_WORKER_INTERVAL" default:"@every 15m"`
	SummarizeWorkerInterval      string `json:"summarize_worker_interval" envconfig:"SUMMARIZE_WORKER_INTERVAL" default:"@every 20m"`
	PurgeWorkerInterval          string `json:"purge_worker_interval" envconfig:"PURGE_WORKER_INTERVAL" default:"@every 1h"`
	DefaultBatchSize             int64  `json:"default_batch_size" envconfig:"DEFAULT_BATCH_SIZE" default:"0"`
	DatabaseDSN                  string `json:"database_dsn" envconfig:"DATABASE_DSN"`
	Debug                        bool   `json:"debug" envconfig:"DEBUG"`
	LogLevel                     string `json:"log_level" envconfig:"LOG_LEVEL" default:"info"`
	LogOutput                    string `json:"log_output" envconfig:"LOG_OUTPUT" default:"stdout"`
	RollbarAccessToken           string `json:"rollbar_access_token" envconfig:"ROLLBAR_ACCESS_TOKEN"`
	RollbarServerRoot            string `json:"rollbar_server_root" envconfig:"ROLLBAR_SERVER_ROOT"`
	IndexerMetricAddr            string `json:"indexer_metric_addr" envconfig:"INDEXER_METRIC_ADDR" default:":8080"`
	ServerMetricAddr             string `json:"server_metric_addr" envconfig:"SERVER_METRIC_ADDR" default:":8090"`
	MetricServerUrl              string `json:"metric_server_url" envconfig:"METRIC_SERVER_URL" default:"/metrics"`
	PurgeSequencesInterval       string `json:"purge_sequences_interval" envconfig:"PURGE_SEQUENCES_INTERVAL" default:"26 hours"`
	PurgeHourlySummariesInterval string `json:"purge_hourly_summaries_interval" envconfig:"PURGE_HOURLY_SUMMARIES_INTERVAL" default:"26h"`
	IndexerConfigFile            string `json:"indexer_config_file" envconfig:"INDEXER_CONFIG_FILE" default:"indexer_config.json"`
}

// Validate returns an error if config is invalid
func (c *Config) Validate() error {
	if c.ProxyUrl == "" {
		return errEndpointRequired
	}

	if c.DatabaseDSN == "" {
		return errDatabaseRequired
	}

	if c.IndexWorkerInterval == "" {
		return errIndexWorkerIntervalRequired
	}

	return nil
}

// IsDevelopment returns true if app is in dev mode
func (c *Config) IsDevelopment() bool {
	return c.AppEnv == modeDevelopment
}

// IsProduction returns true if app is in production mode
func (c *Config) IsProduction() bool {
	return c.AppEnv == modeProduction
}

// ListenAddr returns a full listen address and port
func (c *Config) ListenAddr() string {
	return fmt.Sprintf("%s:%d", c.ServerAddr, c.ServerPort)
}

// New returns a new config
func New() *Config {
	return &Config{}
}

// FromFile reads the config from a file
func FromFile(path string, config *Config) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, config)
}

// FromEnv reads the config from environment variables
func FromEnv(config *Config) error {
	return envconfig.Process("", config)
}
