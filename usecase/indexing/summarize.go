package indexing

import (
	"context"
	"fmt"
	"time"

	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/indexer"
	"github.com/figment-networks/polkadothub-indexer/metric"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
)

type summarizeUseCase struct {
	cfg *config.Config

	blockSeqDb     store.BlockSeq
	blockSummaryDb store.BlockSummary

	validatorEraSeqDb     store.ValidatorEraSeq
	validatorSessionSeqDb store.ValidatorSessionSeq
	validatorSummaryDb    store.ValidatorSummary
}

func NewSummarizeUseCase(cfg *config.Config, blockSeqDb store.BlockSeq, blockSummaryDb store.BlockSummary, validatorEraSeqDb store.ValidatorEraSeq,
	validatorSessionSeqDb store.ValidatorSessionSeq, validatorSummaryDb store.ValidatorSummary,
) *summarizeUseCase {
	return &summarizeUseCase{
		cfg: cfg,

		blockSeqDb:            blockSeqDb,
		blockSummaryDb:        blockSummaryDb,
		validatorEraSeqDb:     validatorEraSeqDb,
		validatorSessionSeqDb: validatorSessionSeqDb,
		validatorSummaryDb:    validatorSummaryDb,
	}
}

func (uc *summarizeUseCase) Execute(ctx context.Context) error {
	defer metric.LogUseCaseDuration(time.Now(), "summarize")

	configParser, err := indexer.NewConfigParser(uc.cfg.IndexerConfigFile)
	if err != nil {
		return err
	}
	currentIndexVersion := configParser.GetCurrentVersionId()

	if err := uc.summarizeBlockSeq(types.IntervalHourly, currentIndexVersion); err != nil {
		return err
	}

	if err := uc.summarizeBlockSeq(types.IntervalDaily, currentIndexVersion); err != nil {
		return err
	}

	if err := uc.summarizeValidatorSeq(types.IntervalHourly, currentIndexVersion); err != nil {
		return err
	}

	if err := uc.summarizeValidatorSeq(types.IntervalDaily, currentIndexVersion); err != nil {
		return err
	}

	return nil
}

func (uc *summarizeUseCase) summarizeBlockSeq(interval types.SummaryInterval, currentIndexVersion int64) error {
	logger.Info(fmt.Sprintf("summarizing block sequences... [interval=%s]", interval))

	activityPeriods, err := uc.blockSummaryDb.FindActivityPeriods(interval, currentIndexVersion)
	if err != nil {
		return err
	}

	rawSummaryItems, err := uc.blockSeqDb.Summarize(interval, activityPeriods)
	if err != nil {
		return err
	}

	var newModels []model.BlockSummary
	var existingModels []model.BlockSummary
	for _, rawSummary := range rawSummaryItems {
		summary := &model.Summary{
			TimeInterval: interval,
			TimeBucket:   rawSummary.TimeBucket,
			IndexVersion: currentIndexVersion,
		}
		query := model.BlockSummary{
			Summary: summary,
		}

		existingBlockSummary, err := uc.blockSummaryDb.Find(&query)
		if err != nil {
			if err == store.ErrNotFound {
				blockSummary := model.BlockSummary{
					Summary: summary,

					Count:        rawSummary.Count,
					BlockTimeAvg: rawSummary.BlockTimeAvg,
				}
				if err := uc.blockSummaryDb.Create(&blockSummary); err != nil {
					return err
				}
				newModels = append(newModels, blockSummary)
			} else {
				return err
			}
		} else {
			existingBlockSummary.Count = rawSummary.Count
			existingBlockSummary.BlockTimeAvg = rawSummary.BlockTimeAvg

			if err := uc.blockSummaryDb.Save(existingBlockSummary); err != nil {
				return err
			}
			existingModels = append(existingModels, *existingBlockSummary)
		}
	}

	logger.Info(fmt.Sprintf("block sequences summarized [created=%d] [updated=%d]", len(newModels), len(existingModels)))

	return nil
}

func (uc *summarizeUseCase) summarizeValidatorSeq(interval types.SummaryInterval, currentIndexVersion int64) error {
	logger.Info(fmt.Sprintf("summarizing validator sequences... [interval=%s]", interval))

	activityPeriods, err := uc.validatorSummaryDb.FindActivityPeriods(interval, currentIndexVersion)
	if err != nil {
		return err
	}

	rawSessionSummaryItems, err := uc.validatorSessionSeqDb.Summarize(interval, activityPeriods)
	if err != nil {
		return err
	}

	rawEraSummaryItems, err := uc.validatorEraSeqDb.Summarize(interval, activityPeriods)
	if err != nil {
		return err
	}

	var newModels []model.ValidatorSummary
	var existingModels []model.ValidatorSummary
	for _, rawSessionSummaryItem := range rawSessionSummaryItems {
		summary := &model.Summary{
			TimeInterval: interval,
			TimeBucket:   rawSessionSummaryItem.TimeBucket,
			IndexVersion: currentIndexVersion,
		}
		query := model.ValidatorSummary{
			Summary: summary,

			StashAccount: rawSessionSummaryItem.StashAccount,
		}

		var rawEraSeqSummary model.ValidatorEraSeqSummary
		for _, rawEraSummaryItem := range rawEraSummaryItems {
			if rawEraSummaryItem.StashAccount == rawSessionSummaryItem.StashAccount && rawEraSummaryItem.TimeBucket == rawSessionSummaryItem.TimeBucket {
				rawEraSeqSummary = rawEraSummaryItem
			}
		}

		existingValidatorSummary, err := uc.validatorSummaryDb.Find(&query)
		if err != nil {
			if err == store.ErrNotFound {
				validatorSummary := model.ValidatorSummary{
					Summary: summary,

					StashAccount: rawSessionSummaryItem.StashAccount,

					TotalStakeAvg:   rawEraSeqSummary.TotalStakeAvg,
					TotalStakeMin:   rawEraSeqSummary.TotalStakeMin,
					TotalStakeMax:   rawEraSeqSummary.TotalStakeMax,
					OwnStakeAvg:     rawEraSeqSummary.OwnStakeAvg,
					OwnStakeMin:     rawEraSeqSummary.OwnStakeMin,
					OwnStakeMax:     rawEraSeqSummary.OwnStakeMax,
					StakersStakeAvg: rawEraSeqSummary.StakersStakeAvg,
					StakersStakeMin: rawEraSeqSummary.StakersStakeMin,
					StakersStakeMax: rawEraSeqSummary.StakersStakeMax,
					RewardPointsAvg: rawEraSeqSummary.RewardPointsAvg,
					RewardPointsMin: rawEraSeqSummary.RewardPointsMin,
					RewardPointsMax: rawEraSeqSummary.RewardPointsMax,
					CommissionAvg:   rawEraSeqSummary.CommissionAvg,
					CommissionMin:   rawEraSeqSummary.CommissionMin,
					CommissionMax:   rawEraSeqSummary.CommissionMax,
					StakersCountAvg: rawEraSeqSummary.StakersCountAvg,
					StakersCountMin: rawEraSeqSummary.StakersCountMin,
					StakersCountMax: rawEraSeqSummary.StakersCountMax,

					UptimeAvg: rawSessionSummaryItem.UptimeAvg,
					UptimeMin: rawSessionSummaryItem.UptimeMin,
					UptimeMax: rawSessionSummaryItem.UptimeMax,
				}

				if err := uc.validatorSummaryDb.Create(&validatorSummary); err != nil {
					return err
				}
				newModels = append(newModels, validatorSummary)
			} else {
				return err
			}
		} else {
			existingValidatorSummary.TotalStakeAvg = rawEraSeqSummary.TotalStakeAvg
			existingValidatorSummary.TotalStakeMin = rawEraSeqSummary.TotalStakeMin
			existingValidatorSummary.TotalStakeMax = rawEraSeqSummary.TotalStakeMax
			existingValidatorSummary.OwnStakeAvg = rawEraSeqSummary.OwnStakeAvg
			existingValidatorSummary.OwnStakeMin = rawEraSeqSummary.OwnStakeMin
			existingValidatorSummary.OwnStakeMax = rawEraSeqSummary.OwnStakeMax
			existingValidatorSummary.StakersStakeAvg = rawEraSeqSummary.StakersStakeAvg
			existingValidatorSummary.StakersStakeMin = rawEraSeqSummary.StakersStakeMin
			existingValidatorSummary.StakersStakeMax = rawEraSeqSummary.StakersStakeMax
			existingValidatorSummary.RewardPointsAvg = rawEraSeqSummary.RewardPointsAvg
			existingValidatorSummary.RewardPointsMin = rawEraSeqSummary.RewardPointsMin
			existingValidatorSummary.RewardPointsMax = rawEraSeqSummary.RewardPointsMax
			existingValidatorSummary.CommissionAvg = rawEraSeqSummary.CommissionAvg
			existingValidatorSummary.CommissionMin = rawEraSeqSummary.CommissionMin
			existingValidatorSummary.CommissionMax = rawEraSeqSummary.CommissionMax
			existingValidatorSummary.StakersCountAvg = rawEraSeqSummary.StakersCountAvg
			existingValidatorSummary.StakersCountMin = rawEraSeqSummary.StakersCountMin
			existingValidatorSummary.StakersCountMax = rawEraSeqSummary.StakersCountMax

			existingValidatorSummary.UptimeAvg = rawSessionSummaryItem.UptimeAvg
			existingValidatorSummary.UptimeMin = rawSessionSummaryItem.UptimeMin
			existingValidatorSummary.UptimeMax = rawSessionSummaryItem.UptimeMax

			if err := uc.validatorSummaryDb.Save(existingValidatorSummary); err != nil {
				return err
			}
			existingModels = append(existingModels, *existingValidatorSummary)
		}
	}

	logger.Info(fmt.Sprintf("validator sequences summarized [created=%d] [updated=%d]", len(newModels), len(existingModels)))

	return nil
}
