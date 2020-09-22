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

	workerHandlers := usecase.NewWorkerHandlers(cfg, client, db.GetAccountEraSeq(), db.GetBlockSeq(), db.GetBlockSummary(), db.GetDatabase(), db.GetEventSeq(),
		db.GetReports(), db.GetSyncables(), db.GetValidatorAgg(), db.GetValidatorEraSeq(), db.GetValidatorSessionSeq(), db.GetValidatorSummary(),
	)

	w, err := worker.New(cfg, workerHandlers)
	if err != nil {
		return err
	}

	return w.Start()
}
