package validator

import (
	"context"

	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/indexer"
	"github.com/figment-networks/polkadothub-indexer/store"

	"github.com/pkg/errors"
)

type getByHeightUseCase struct {
	cfg    *config.Config
	client *client.Client

	accountDb     store.Accounts
	blockDb       store.Blocks
	databaseDb    store.Database
	eventDb       store.Events
	reportDb      store.Reports
	syncableDb    store.Syncables
	transactionDb store.Transactions
	validatorDb   store.Validators
}

func NewGetByHeightUseCase(cfg *config.Config, cli *client.Client, accountDb store.Accounts, blockDb store.Blocks, databaseDb store.Database, eventDb store.Events, reportDb store.Reports,
	syncableDb store.Syncables, transactionDb store.Transactions, validatorDb store.Validators) *getByHeightUseCase {
	return &getByHeightUseCase{
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

func (uc *getByHeightUseCase) Execute(height *int64) (SeqListView, error) {
	// Get last indexed height
	mostRecentSynced, err := uc.syncableDb.FindMostRecent()
	if err != nil {
		return SeqListView{}, err
	}
	lastH := mostRecentSynced.Height

	// Show last synced height, if not provided
	if height == nil {
		height = &lastH
	}

	if *height > lastH {
		return SeqListView{}, errors.New("height is not indexed yet")
	}

	sessionSeqs, err := uc.validatorDb.FindSessionSeqsByHeight(*height)
	if len(sessionSeqs) == 0 || err != nil {
		syncable, err := uc.syncableDb.FindLastInSessionForHeight(*height)
		if err != nil {
			return SeqListView{}, err
		}

		indexingPipeline, err := indexer.NewPipeline(uc.cfg, uc.client, uc.accountDb, uc.blockDb, uc.databaseDb, uc.eventDb, uc.reportDb, uc.syncableDb, uc.transactionDb, uc.validatorDb)
		if err != nil {
			return SeqListView{}, err
		}

		ctx := context.Background()
		payload, err := indexingPipeline.Run(ctx, indexer.RunConfig{
			Height:           syncable.Height,
			DesiredTargetIDs: []int64{indexer.TargetIndexValidatorSessionSequences},
			Dry:              true,
		})
		if err != nil {
			return SeqListView{}, err
		}

		sessionSeqs = append(payload.NewValidatorSessionSequences, payload.UpdatedValidatorSessionSequences...)
	}

	eraSeqs, err := uc.validatorDb.FindEraSeqsByHeight(*height)
	if err != nil && err != store.ErrNotFound {
		return SeqListView{}, err
	}

	return ToSeqListView(sessionSeqs, eraSeqs), nil
}
