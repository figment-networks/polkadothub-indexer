package indexing

import (
	"context"
	"fmt"

	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
)

type StartCmdHandler struct {
	cfg    *config.Config
	client *client.Client

	useCase *startUseCase

	accountDb     store.Accounts
	blockDb       store.Blocks
	databaseDb    store.Database
	eventDb       store.Events
	reportDb      store.Reports
	syncableDb    store.Syncables
	transactionDb store.Transactions
	validatorDb   store.Validators
}

func NewStartCmdHandler(cfg *config.Config, cli *client.Client, accountDb store.Accounts, blockDb store.Blocks, databaseDb store.Database, eventDb store.Events, reportDb store.Reports,
	syncableDb store.Syncables, transactionDb store.Transactions, validatorDb store.Validators,
) *StartCmdHandler {
	return &StartCmdHandler{
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

func (h *StartCmdHandler) Handle(ctx context.Context, batchSize int64) {
	logger.Info(fmt.Sprintf("running indexer use case [handler=cmd] [batchSize=%d]", batchSize))

	err := h.getUseCase().Execute(ctx, batchSize)
	if err != nil {
		logger.Error(err)
		return
	}
}

func (h *StartCmdHandler) getUseCase() *startUseCase {
	if h.useCase == nil {
		return NewStartUseCase(h.cfg, h.client, h.accountDb, h.blockDb, h.databaseDb, h.eventDb, h.reportDb, h.syncableDb, h.transactionDb, h.validatorDb)
	}
	return h.useCase
}
