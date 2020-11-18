package validator

import (
	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/usecase/http"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

var (
	_ types.HttpHandler = (*getByHeightHttpHandler)(nil)
)

type getByHeightHttpHandler struct {
	cfg    *config.Config
	client *client.Client

	useCase *getByHeightUseCase

	accountDb     store.Accounts
	blockDb       store.Blocks
	databaseDb    store.Database
	eventDb       store.Events
	reportDb      store.Reports
	syncableDb    store.Syncables
	systemEventDb store.SystemEvents
	transactionDb store.Transactions
	validatorDb   store.Validators
}

func NewGetByHeightHttpHandler(cfg *config.Config, cli *client.Client, accountDb store.Accounts, blockDb store.Blocks, databaseDb store.Database, eventDb store.Events, reportDb store.Reports,
	syncableDb store.Syncables, systemEventDb store.SystemEvents, transactionDb store.Transactions, validatorDb store.Validators,
) *getByHeightHttpHandler {
	return &getByHeightHttpHandler{
		cfg:    cfg,
		client: cli,

		accountDb:     accountDb,
		blockDb:       blockDb,
		databaseDb:    databaseDb,
		eventDb:       eventDb,
		reportDb:      reportDb,
		syncableDb:    syncableDb,
		systemEventDb: systemEventDb,
		transactionDb: transactionDb,
		validatorDb:   validatorDb,
	}
}

type GetByHeightRequest struct {
	Height *int64 `form:"height" binding:"-"`
}

func (h *getByHeightHttpHandler) Handle(c *gin.Context) {
	var req GetByHeightRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		logger.Error(err)
		http.BadRequest(c, errors.New("invalid height"))
		return
	}

	ds, err := h.getUseCase().Execute(req.Height)
	if err != nil {
		logger.Error(err)
		http.ServerError(c, err)
		return
	}

	http.JsonOK(c, ds)
}

func (h *getByHeightHttpHandler) getUseCase() *getByHeightUseCase {
	if h.useCase == nil {
		return NewGetByHeightUseCase(h.cfg, h.client, h.accountDb, h.blockDb, h.databaseDb, h.eventDb, h.reportDb, h.syncableDb, h.systemEventDb, h.transactionDb, h.validatorDb)
	}
	return h.useCase
}
