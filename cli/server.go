package cli

import (
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/server"
	"github.com/figment-networks/polkadothub-indexer/usecase"
)

func startServer(cfg *config.Config) error {
	client, err := initClient(cfg)
	if err != nil {
		return err
	}
	defer client.Close()
	db, err := initStore(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	httpHandlers, err := usecase.NewHttpHandlers(&usecase.HttpHandlerParams{
		Client:                client,
		AccountEraSeqDb:       db.GetAccountEraSeq(),
		BlockSeqDb:            db.GetBlockSeq(),
		BlockSummaryDb:        db.GetBlockSummary(),
		EventSeqDb:            db.GetEventSeq(),
		SyncablesDb:           db.GetSyncables(),
		ValidatorAggDb:        db.GetValidatorAgg(),
		ValidatorEraSeqDb:     db.GetValidatorEraSeq(),
		ValidatorSessionSeqDb: db.GetValidatorSessionSeq(),
		ValidatorSummaryDb:    db.GetValidatorSummary(),
	})
	if err != nil {
		return err
	}

	a := server.New(cfg, httpHandlers)
	if err := a.Start(cfg.ListenAddr()); err != nil {
		return err
	}
	return nil
}
