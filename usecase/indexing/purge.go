package indexing

import (
	"context"
	"fmt"
	"time"

	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/indexer"
	"github.com/figment-networks/polkadothub-indexer/metric"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
	"github.com/pkg/errors"
)

var (
	ErrPurgingDisabled = errors.New("purging disabled")
)

type purgeUseCase struct {
	cfg *config.Config

	blockDb     store.Blocks
	validatorDb store.Validators
}

func NewPurgeUseCase(cfg *config.Config, blockDb store.Blocks, validatorDb store.Validators) *purgeUseCase {
	return &purgeUseCase{
		cfg: cfg,

		blockDb:     blockDb,
		validatorDb: validatorDb,
	}
}

func (uc *purgeUseCase) Execute(ctx context.Context) error {
	defer metric.LogUseCaseDuration(time.Now(), "purge")

	configParser, err := indexer.NewConfigParser(uc.cfg.IndexerConfigFile)
	if err != nil {
		return err
	}
	currentIndexVersion := configParser.GetCurrentVersionId()

	if err := uc.purgeBlocks(currentIndexVersion); err != nil {
		return err
	}

	if err := uc.purgeValidators(currentIndexVersion); err != nil {
		return err
	}

	return nil
}

func (uc *purgeUseCase) purgeBlocks(currentIndexVersion int64) error {
	if err := uc.purgeBlockSequences(currentIndexVersion); uc.checkErr(err) {
		return err
	}
	if err := uc.purgeBlockSummaries(types.IntervalHourly, uc.cfg.PurgeHourlySummariesInterval); uc.checkErr(err) {
		return err
	}
	return nil
}

func (uc *purgeUseCase) purgeValidators(currentIndexVersion int64) error {
	if err := uc.purgeValidatorSessionSequences(currentIndexVersion); uc.checkErr(err) {
		return err
	}

	if err := uc.purgeValidatorSummaries(types.IntervalHourly, uc.cfg.PurgeHourlySummariesInterval); uc.checkErr(err) {
		return err
	}

	// never purge validator era sequences because refetching data is very expensive
	return nil
}

func (uc *purgeUseCase) purgeBlockSequences(currentIndexVersion int64) error {
	blockSeq, err := uc.blockDb.FindMostRecentSeq()
	if err != nil {
		return err
	}
	lastSeqTime := blockSeq.Time.Time

	duration, err := uc.parseDuration(uc.cfg.PurgeSequencesInterval)
	if err != nil {
		if err == ErrPurgingDisabled {
			logger.Info("purging block sequences disabled. Purge interval set to 0.")
		}
		return err
	}

	purgeThresholdFromLastSeq := lastSeqTime.Add(-*duration)

	activityPeriods, err := uc.blockDb.FindActivityPeriods(types.IntervalDaily, currentIndexVersion)
	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("purging summarized block sequences... [older than=%s]", purgeThresholdFromLastSeq))

	deletedCount, err := uc.blockDb.DeleteSeqOlderThan(purgeThresholdFromLastSeq, activityPeriods)
	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("%d block sequences purged", *deletedCount))

	return nil
}

func (uc *purgeUseCase) purgeBlockSummaries(interval types.SummaryInterval, purgeInterval string) error {
	blockSummary, err := uc.blockDb.FindMostRecentByInterval(interval)
	if err != nil {
		return err
	}
	lastSummaryTimeBucket := blockSummary.TimeBucket.Time

	duration, err := uc.parseDuration(purgeInterval)
	if err != nil {
		if err == ErrPurgingDisabled {
			logger.Info(fmt.Sprintf("purging block summaries disabled [interval=%s] [purge_interval=%s]", interval, purgeInterval))
		}
		return err
	}

	purgeThreshold := lastSummaryTimeBucket.Add(-*duration)

	logger.Info(fmt.Sprintf("purging block summaries... [interval=%s] [older than=%s]", interval, purgeThreshold))

	deletedCount, err := uc.blockDb.DeleteOlderThan(interval, purgeThreshold)
	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("%d block summaries purged [interval=%s]", *deletedCount, interval))

	return nil
}

func (uc *purgeUseCase) purgeValidatorSessionSequences(currentIndexVersion int64) error {
	validatorSeq, err := uc.validatorDb.FindMostRecentSessionSeq()
	if err != nil {
		return err
	}
	lastSeqTime := validatorSeq.Time.Time

	validatorSummary, err := uc.validatorDb.FindMostRecentSummary()
	if err != nil {
		return err
	}
	lastSummaryTimeBucket := validatorSummary.TimeBucket.Time

	duration, err := uc.parseDuration(uc.cfg.PurgeSequencesInterval)
	if err != nil {
		if err == ErrPurgingDisabled {
			logger.Info("purging validator sequences disabled. Purge interval set to 0.")
		}
		return err
	}

	purgeThresholdFromConfig := lastSeqTime.Add(-*duration)

	var purgeThreshold time.Time
	if purgeThresholdFromConfig.Before(lastSummaryTimeBucket) {
		purgeThreshold = purgeThresholdFromConfig
	} else {
		purgeThreshold = lastSummaryTimeBucket
	}

	logger.Info(fmt.Sprintf("purging validator session sequences... [older than=%s]", purgeThreshold))

	deletedCount, err := uc.validatorDb.DeleteSessionSeqsOlderThan(purgeThreshold)
	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("%d validator session sequences purged", *deletedCount))

	return nil
}

func (uc *purgeUseCase) purgeValidatorSummaries(interval types.SummaryInterval, purgeInterval string) error {
	blockSummary, err := uc.validatorDb.FindMostRecentByInterval(interval)
	if err != nil {
		return err
	}
	lastSummaryTimeBucket := blockSummary.TimeBucket.Time

	duration, err := uc.parseDuration(purgeInterval)
	if err != nil {
		if err == ErrPurgingDisabled {
			logger.Info(fmt.Sprintf("purging validator summaries disabled [interval=%s] [purge_interval=%s]", interval, purgeInterval))
		}
		return err
	}

	purgeThreshold := lastSummaryTimeBucket.Add(-*duration)

	logger.Info(fmt.Sprintf("purging validator summaries... [interval=%s] [older than=%s]", interval, purgeThreshold))

	deletedCount, err := uc.validatorDb.DeleteSummaryOlderThan(interval, purgeThreshold)
	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("%d validator summaries purged [interval=%s]", *deletedCount, interval))

	return nil
}

func (uc *purgeUseCase) parseDuration(interval string) (*time.Duration, error) {
	duration, err := time.ParseDuration(interval)
	if err != nil {
		return nil, err
	}

	if duration == 0 {
		return nil, ErrPurgingDisabled
	}
	return &duration, nil
}

func (uc *purgeUseCase) checkErr(err error) bool {
	return err != nil && err != ErrPurgingDisabled && err != store.ErrNotFound
}
