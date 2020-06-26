package indexer

import (
	"github.com/figment-networks/indexing-engine/pipeline"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
)

func NewLogger() pipeline.Logger {
	return &idxLogger{}
}

type idxLogger struct {}

func (l idxLogger) Info(msg string) {
	logger.Info(msg)
}

func (l idxLogger) Debug(msg string) {
	logger.Debug(msg)
}
