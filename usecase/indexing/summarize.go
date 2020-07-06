package indexing

import (
	"context"
	"fmt"
	"github.com/figment-networks/polkadothub-indexer/indexer"
	"time"

	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/metric"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
)

type summarizeUseCase struct {
	cfg *config.Config
	db  *store.Store
}

func NewSummarizeUseCase(cfg *config.Config, db *store.Store) *summarizeUseCase {
	return &summarizeUseCase{
		cfg: cfg,
		db:  db,
	}
}

func (uc *summarizeUseCase) Execute(ctx context.Context) error {
	defer metric.LogUseCaseDuration(time.Now(), "summarize")

	targetsReader, err := indexer.NewTargetsReader(uc.cfg.IndexerTargetsFile)
	if err != nil {
		return err
	}
	currentIndexVersion := targetsReader.GetCurrentVersion()

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

	activityPeriods, err := uc.db.BlockSummary.FindActivityPeriods(interval, currentIndexVersion)
	if err != nil {
		return err
	}

	rawSummaryItems, err := uc.db.BlockSeq.Summarize(interval, activityPeriods)
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

		existingBlockSummary, err := uc.db.BlockSummary.Find(&query)
		if err != nil {
			if err == store.ErrNotFound {
				blockSummary := model.BlockSummary{
					Summary: summary,

					Count:        rawSummary.Count,
					BlockTimeAvg: rawSummary.BlockTimeAvg,
				}
				if err := uc.db.BlockSummary.Create(&blockSummary); err != nil {
					return err
				}
				newModels = append(newModels, blockSummary)
			} else {
				return err
			}
		} else {
			existingBlockSummary.Count = rawSummary.Count
			existingBlockSummary.BlockTimeAvg = rawSummary.BlockTimeAvg

			if err := uc.db.BlockSummary.Save(existingBlockSummary); err != nil {
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

	activityPeriods, err := uc.db.ValidatorSummary.FindActivityPeriods(interval, currentIndexVersion)
	if err != nil {
		return err
	}

	rawSessionSummaryItems, err := uc.db.ValidatorSessionSeq.Summarize(interval, activityPeriods)
	if err != nil {
		return err
	}

	rawEraSummaryItems, err := uc.db.ValidatorEraSeq.Summarize(interval, activityPeriods)
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

		var rawEraSeqSummary store.ValidatorEraSeqSummary
		for _, rawEraSummaryItem := range rawEraSummaryItems {
			if rawEraSummaryItem.TimeBucket == rawSessionSummaryItem.TimeBucket {
				rawEraSeqSummary = rawEraSummaryItem
			}
		}

		existingValidatorSummary, err := uc.db.ValidatorSummary.Find(&query)
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

				if err := uc.db.ValidatorSummary.Create(&validatorSummary); err != nil {
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

			if err := uc.db.ValidatorSummary.Save(existingValidatorSummary); err != nil {
				return err
			}
			existingModels = append(existingModels, *existingValidatorSummary)
		}
	}

	logger.Info(fmt.Sprintf("validator sequences summarized [created=%d] [updated=%d]", len(newModels), len(existingModels)))

	return nil
}
