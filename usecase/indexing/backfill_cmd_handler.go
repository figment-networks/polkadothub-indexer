package indexing

import (
	"context"

	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
)

type BackfillCmdHandler struct {
	cfg    *config.Config
	client *client.Client

	useCase *backfillUseCase

	accountDb     store.Accounts
	blockDb       store.Blocks
	databaseDb    store.Database
	eventDb       store.Events
	reportDb      store.Reports
	syncableDb    store.Syncables
	transactionDb store.Transactions
	validatorDb   store.Validators
}

func NewBackfillCmdHandler(cfg *config.Config, cli *client.Client, accountDb store.Accounts, blockDb store.Blocks, databaseDb store.Database, eventDb store.Events, reportDb store.Reports,
	syncableDb store.Syncables, transactionDb store.Transactions, validatorDb store.Validators,
) *BackfillCmdHandler {
	return &BackfillCmdHandler{
		cfg:    cfg,
		client: cli,

		accountDb:     accountDb,
		blockDb:       blockDb,
		databaseDb:    databaseDb,
		eventDb:       eventDb,
		reportDb:      reportDb,
		syncableDb:    syncableDb,
		transactionDb: transactionDb,
		validatorDb:   validatorDb,
	}
}

func (h *BackfillCmdHandler) Handle(ctx context.Context, parallel bool, force bool, targetIds []int64) {
	logger.Info("running backfill use case [handler=cmd]")

	useCaseConfig := BackfillUseCaseConfig{
		Parallel:  parallel,
		Force:     force,
		TargetIds: targetIds,
	}
	err := h.getUseCase().Execute(ctx, useCaseConfig)
	if err != nil {
		logger.Error(err)
		return
	}
}

func (h *BackfillCmdHandler) getUseCase() *backfillUseCase {
	if h.useCase == nil {
		return NewBackfillUseCase(h.cfg, h.client, h.accountDb, h.blockDb, h.databaseDb, h.eventDb, h.reportDb, h.syncableDb, h.transactionDb, h.validatorDb)
	}
	return h.useCase
}
