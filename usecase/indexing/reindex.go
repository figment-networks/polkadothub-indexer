package indexing

import (
	"context"
	"errors"

	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/indexer"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/store"
)

var (
	ErrReindexRunning = errors.New("reindex already running (use force flag to override it)")
)

type reindexUseCase struct {
	cfg    *config.Config
	client *client.Client

	accountDb     store.Accounts
	blockDb       store.Blocks
	databaseDb    store.Database
	eventDb       store.Events
	reportDb      store.Reports
	rewardDb      store.Rewards
	syncableDb    store.Syncables
	systemEventDb store.SystemEvents
	transactionDb store.Transactions
	validatorDb   store.Validators
}

func NewReindexUseCase(cfg *config.Config, cli *client.Client, accountDb store.Accounts, blockDb store.Blocks, databaseDb store.Database, eventDb store.Events,
	reportDb store.Reports, rewardDb store.Rewards, syncableDb store.Syncables, systemEventDb store.SystemEvents, transactionDb store.Transactions, validatorDb store.Validators,
) *reindexUseCase {
	return &reindexUseCase{
		cfg:    cfg,
		client: cli,

		accountDb:     accountDb,
		blockDb:       blockDb,
		databaseDb:    databaseDb,
		eventDb:       eventDb,
		reportDb:      reportDb,
		rewardDb:      rewardDb,
		syncableDb:    syncableDb,
		systemEventDb: systemEventDb,
		transactionDb: transactionDb,
		validatorDb:   validatorDb,
	}
}

type ReindexUseCaseConfig struct {
	Parallel      bool
	Force         bool
	TargetIds     []int64
	LastInSession bool
	LastInEra     bool
	TrxKinds      []model.TransactionKind
	StartHeight   int64
	EndHeight     int64
}

func (uc *reindexUseCase) Execute(ctx context.Context, useCaseConfig ReindexUseCaseConfig) error {
	if err := uc.canExecute(useCaseConfig.Force); err != nil {
		return err
	}

	indexingPipeline, err := indexer.NewPipeline(uc.cfg, uc.client, uc.accountDb, uc.blockDb, uc.databaseDb, uc.eventDb, uc.reportDb, uc.rewardDb, uc.syncableDb, uc.systemEventDb, uc.transactionDb, uc.validatorDb)
	if err != nil {
		return err
	}

	return indexingPipeline.Reindex(ctx, indexer.ReindexConfig{
		Parallel:      useCaseConfig.Parallel,
		Force:         useCaseConfig.Force,
		TargetIds:     useCaseConfig.TargetIds,
		LastInSession: useCaseConfig.LastInSession,
		LastInEra:     useCaseConfig.LastInEra,
		TrxKinds:      useCaseConfig.TrxKinds,
		StartHeight:   useCaseConfig.StartHeight,
		EndHeight:     useCaseConfig.EndHeight,
	})
}

// canExecute checks if reindex is already running
// if is it running and force is not set as true then we skip indexing
func (uc *reindexUseCase) canExecute(force bool) error {
	if force {
		return nil
	}

	if _, err := uc.reportDb.FindNotCompletedByKind(model.ReportKindSequentialReindex, model.ReportKindParallelReindex); err != nil {
		if err == store.ErrNotFound {
			return nil
		}
		return err
	}
	return ErrReindexRunning
}
