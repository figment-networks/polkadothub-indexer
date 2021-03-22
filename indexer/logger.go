package indexer

import (
	"github.com/figment-networks/indexing-engine/pipeline"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
	"go.uber.org/zap"
)

func NewLogger() pipeline.Logger {
	return &idxLogger{}
}

type idxLogger struct{}

func (l idxLogger) Info(msg string) {
	logger.Info(msg)
}

func (l idxLogger) Debug(msg string) {
	logger.Debug(msg)
}

func (l idxLogger) Error(err error, tags ...zap.Field) {
	logger.Error(err, tags...)
}
