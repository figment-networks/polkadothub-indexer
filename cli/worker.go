package cli

import (
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/usecase"
	"github.com/figment-networks/polkadothub-indexer/worker"
)

func startWorker(cfg *config.Config) error {
	db, err := initPostgres(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	client, err := initClient(cfg)
	if err != nil {
		return err
	}
	defer client.Close()

	workerHandlers := usecase.NewWorkerHandlers(cfg, client, db.GetAccounts(), db.GetBlocks(), db.GetDatabase(), db.GetEvents(), db.GetReports(),
		db.GetRewards(), db.GetSyncables(), db.GetSystemEvents(), db.GetTransactions(), db.GetValidators(),
	)

	w, err := worker.New(cfg, workerHandlers)
	if err != nil {
		return err
	}

	return w.Start()
}
