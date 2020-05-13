package config

import (
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/spf13/viper"
)

const (
	// http API port
	appPort       = "PORT"
	// polkadothub-proxy url
	proxyUrl      = "PROXY_URL"
	// log level
	logLevel      = "LOG_LEVEL"
	// where to output logs
	logOutput     = "LOG_OUTPUT"
	// current go environment
	goEnvironment = "GO_ENVIRONMENT"

	// CLI
	// command line argument to be used for passing batch size
	batchSize = "CLI_BATCH_SIZE_ARG"

	// Job
	// Size of the batch for processing pipeline
	pipelineBatchSize  = "BLOCK_SYNC_BATCH_SIZE"
	// How often processing job should be run
	processingInterval = "PROCESSING_INTERVAL"
	// How often cleanup job should be run
	cleanupInterval    = "CLEANUP_INTERVAL"
	// Height of the first block
	firstBlockHeight   = "FIRST_BLOCK_HEIGHT"
	// How many syncables do we want to keep in our database before deleting them
	cleanupThreshold   = "CLEANUP_THRESHOLD"

	// Database
	dbUser        = "DB_USER"
	dbPassword    = "DB_PASSWORD"
	dbHost        = "DB_HOST"
	dbName        = "DB_NAME"
	dbDetailedLog = "DB_DETAILED_LOG"
	dbSSLMode     = "DB_SSL_MODE"

	production  = "production"
	development = "development"

	// Rollbar
	rollbarAccessToken = "ROLLBAR_ACCESS_TOKEN"
)

var (
	defaultAppPort       = "8081"
	defaultProxyUrl      = "localhost:50051"
	defaultLogLevel      = "info"
	defaultLogOutput     = "stdout"
	defaultGoEnvironment = development

	defaultBatchSize = "batchSize"

	defaultPipelineBatchSize  int64 = 2
	defaultProcessingInterval       = "@every 10s"
	defaultCleanupInterval          = "@every 10m"
	defaultFirstBlockHeight   int64 = 1
	defaultCleanupThreshold   int64 = 1000

	defaultDbUser        = "postgres"
	defaultDbPassword    = "password"
	defaultDbHost        = "timedb"
	defaultDbName        = "app"
	defaultSSLMode       = "disable"
	dbDefaultDetailedLog = false
)

func init() {
	viper.SetDefault(appPort, defaultAppPort)
	viper.SetDefault(proxyUrl, defaultProxyUrl)
	viper.SetDefault(logLevel, defaultLogLevel)
	viper.SetDefault(logOutput, defaultLogOutput)
	viper.SetDefault(goEnvironment, defaultGoEnvironment)

	viper.SetDefault(batchSize, defaultBatchSize)

	viper.SetDefault(pipelineBatchSize, defaultPipelineBatchSize)
	viper.SetDefault(processingInterval, defaultProcessingInterval)
	viper.SetDefault(cleanupInterval, defaultCleanupInterval)
	viper.SetDefault(firstBlockHeight, defaultFirstBlockHeight)
	viper.SetDefault(cleanupThreshold, defaultCleanupThreshold)

	viper.SetDefault(dbUser, defaultDbUser)
	viper.SetDefault(dbPassword, defaultDbPassword)
	viper.SetDefault(dbHost, defaultDbHost)
	viper.SetDefault(dbName, defaultDbName)
	viper.SetDefault(dbSSLMode, defaultSSLMode)
	viper.SetDefault(dbDetailedLog, dbDefaultDetailedLog)

	viper.BindEnv(appPort)
	viper.BindEnv(proxyUrl)
	viper.BindEnv(logLevel)
	viper.BindEnv(logOutput)
	viper.BindEnv(goEnvironment)

	viper.BindEnv(batchSize)

	viper.BindEnv(pipelineBatchSize)
	viper.BindEnv(processingInterval)
	viper.BindEnv(cleanupInterval)
	viper.BindEnv(firstBlockHeight)
	viper.BindEnv(cleanupThreshold)

	viper.BindEnv(dbUser)
	viper.BindEnv(dbPassword)
	viper.BindEnv(dbHost)
	viper.BindEnv(dbName)
	viper.BindEnv(dbSSLMode)
	viper.BindEnv(dbDetailedLog)

	viper.BindEnv(rollbarAccessToken)
}

func AppPort() string {
	return viper.GetString(appPort)
}

func ProxyUrl() string {
	return viper.GetString(proxyUrl)
}

func LogLevel() string {
	return viper.GetString(logLevel)
}

func LogOutput() string {
	return viper.GetString(logOutput)
}

func GoEnvironment() string {
	return viper.GetString(goEnvironment)
}

func BatchSize() string {
	return viper.GetString(batchSize)
}

func PipelineBatchSize() int64 {
	return viper.GetInt64(pipelineBatchSize)
}

func ProcessingInterval() string {
	return viper.GetString(processingInterval)
}

func CleanupInterval() string {
	return viper.GetString(cleanupInterval)
}

func FirstBlockHeight() types.Height {
	return types.Height(viper.GetInt64(firstBlockHeight))
}

func CleanupThreshold() int64 {
	return viper.GetInt64(cleanupThreshold)
}

func DbName() string {
	return viper.GetString(dbName)
}

func DbUser() string {
	return viper.GetString(dbUser)
}

func DbHost() string {
	return viper.GetString(dbHost)
}

func DbPassword() string {
	return viper.GetString(dbPassword)
}

func DbSSLMode() string {
	return viper.GetString(dbSSLMode)
}

func DbDetailedLog() bool {
	return viper.GetBool(dbDetailedLog)
}

func RollbarAccessToken() string {
	return viper.GetString(rollbarAccessToken)
}
