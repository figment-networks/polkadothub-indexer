package cli

import (
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/usecase"
	"github.com/figment-networks/polkadothub-indexer/worker"
)

func startWorker(cfg *config.Config) error {
	db, err := initStore(cfg)
	if err != nil {
		return err
	}
	defer db.Close()
	client, err := initClient(cfg)
	if err != nil {
		return err
	}
	defer client.Close()

	workerHandlers, err := usecase.NewWorkerHandlers(&usecase.WorkerHandlerParams{
		Config:                cfg,
		Client:                client,
		AccountEraSeqDb:       db.GetAccountEraSeq(),
		BlockSeqDb:            db.GetBlockSeq(),
		BlockSummaryDb:        db.GetBlockSummary(),
		DatabaseDb:            db.GetDatabase(),
		EventSeqDb:            db.GetEventSeq(),
		ReportsDb:             db.GetReports(),
		SyncablesDb:           db.GetSyncables(),
		ValidatorAggDb:        db.GetValidatorAgg(),
		ValidatorEraSeqDb:     db.GetValidatorEraSeq(),
		ValidatorSessionSeqDb: db.GetValidatorSessionSeq(),
		ValidatorSummaryDb:    db.GetValidatorSummary(),
	})
	if err != nil {
		return err
	}

	w, err := worker.New(cfg, workerHandlers)
	if err != nil {
		return err
	}

	return w.Start()
}
