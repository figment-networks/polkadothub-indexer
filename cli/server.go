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

	db, err := initPostgres(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	httpHandlers, err := usecase.NewHttpHandlers(client, db.GetAccounts(), db.GetBlocks(), db.GetEvents(), db.GetSyncables(), db.GetValidators())
	if err != nil {
		return err
	}

	a := server.New(cfg, httpHandlers)
	if err := a.Start(cfg.ListenAddr()); err != nil {
		return err
	}
	return nil
}
