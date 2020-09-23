package cli

import (
	"context"
	"fmt"

	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/usecase"
	"github.com/figment-networks/polkadothub-indexer/utils/logger"
	"github.com/pkg/errors"
)

func runCmd(cfg *config.Config, flags Flags) error {
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

	cmdHandlers, err := usecase.NewCmdHandlers(&usecase.CmdHandlerParams{
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

	logger.Info(fmt.Sprintf("executing cmd %s ...", flags.runCommand), logger.Field("app", "cli"))

	ctx := context.Background()
	switch flags.runCommand {
	case "status":
		cmdHandlers.GetStatus.Handle(ctx)
	case "indexer_start":
		cmdHandlers.StartIndexer.Handle(ctx, flags.batchSize)
	case "indexer_backfill":
		cmdHandlers.BackfillIndexer.Handle(ctx, flags.parallel, flags.force, flags.targetIds)
	case "indexer_summarize":
		cmdHandlers.SummarizeIndexer.Handle(ctx)
	case "indexer_purge":
		cmdHandlers.PurgeIndexer.Handle(ctx)
	default:
		return errors.New(fmt.Sprintf("command %s not found", flags.runCommand))
	}
	return nil
}
