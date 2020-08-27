package indexing

import (
	"context"

	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/indexer"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/pkg/errors"
)

var (
	ErrRunningSequentialReindex = errors.New("indexing skipped because sequential reindex hasn't finished yet")
)

type startUseCase struct {
	cfg    *config.Config
	db     *store.Store
	client *client.Client
}

func NewStartUseCase(cfg *config.Config, db *store.Store, c *client.Client) *startUseCase {
	return &startUseCase{
		cfg:    cfg,
		db:     db,
		client: c,
	}
}

func (uc *startUseCase) Execute(ctx context.Context, batchSize int64) error {
	if err := uc.canExecute(); err != nil {
		return err
	}

	indexingPipeline, err := indexer.NewPipeline(uc.cfg, uc.db, uc.client)
	if err != nil {
		return err
	}

	return indexingPipeline.Start(ctx, indexer.IndexConfig{
		BatchSize: batchSize,
	})
}

// canExecute checks if sequential reindex is already running
// if is it running we skip indexing
func (uc *startUseCase) canExecute() error {
	if _, err := uc.db.Reports.FindNotCompletedByKind(model.ReportKindSequentialReindex); err != nil {
		if err == store.ErrNotFound {
			return nil
		}
		return err
	}
	return ErrRunningSequentialReindex
}
