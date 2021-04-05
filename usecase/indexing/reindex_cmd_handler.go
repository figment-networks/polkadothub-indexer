package indexing

import (
	"context"

	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
)

type ReindexCmdHandler struct {
	cfg    *config.Config
	client *client.Client

	useCase *reindexUseCase

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

func NewReindexCmdHandler(cfg *config.Config, cli *client.Client, accountDb store.Accounts, blockDb store.Blocks, databaseDb store.Database, eventDb store.Events, reportDb store.Reports,
	rewardDb store.Rewards, syncableDb store.Syncables, systemEventDb store.SystemEvents, transactionDb store.Transactions, validatorDb store.Validators,
) *ReindexCmdHandler {
	return &ReindexCmdHandler{
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

func (h *ReindexCmdHandler) Handle(ctx context.Context, parallel bool, force bool, targetIds []int64, lastInEra bool, lastInSession bool, trxKinds []model.TransactionKind, startHeight, endHeight int64) {
	logger.Info("running reindex use case [handler=cmd]")

	useCaseConfig := ReindexUseCaseConfig{
		Parallel:      parallel,
		Force:         force,
		TargetIds:     targetIds,
		LastInSession: lastInSession,
		LastInEra:     lastInEra,
		TrxKinds:      trxKinds,
		StartHeight:   startHeight,
		EndHeight:     endHeight,
	}
	err := h.getUseCase().Execute(ctx, useCaseConfig)
	if err != nil {
		logger.Error(err)
		return
	}
}

func (h *ReindexCmdHandler) getUseCase() *reindexUseCase {
	if h.useCase == nil {
		return NewReindexUseCase(h.cfg, h.client, h.accountDb, h.blockDb, h.databaseDb, h.eventDb, h.reportDb, h.rewardDb, h.syncableDb, h.systemEventDb, h.transactionDb, h.validatorDb)
	}
	return h.useCase
}
