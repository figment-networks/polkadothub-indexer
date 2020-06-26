package indexer

import (
	"context"
	"fmt"
	"github.com/figment-networks/indexing-engine/pipeline"
	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/pkg/errors"
)

var (
	_ pipeline.Source = (*backfillSource)(nil)
)

type BackfillSourceConfig struct {
	indexVersion int64
}

func NewBackfillSource(cfg *config.Config, db *store.Store, client *client.Client, sourceCfg *BackfillSourceConfig) (*backfillSource, error) {
	src := &backfillSource{
		cfg:    cfg,
		db:     db,
		client: client,

		sourceCfg: sourceCfg,
	}

	if err := src.init(); err != nil {
		return nil, err
	}

	return src, nil
}

type backfillSource struct {
	cfg    *config.Config
	db     *store.Store
	client *client.Client

	sourceCfg *BackfillSourceConfig

	currentHeight int64
	startHeight   int64
	endHeight     int64
	err           error
}

func (s *backfillSource) Next(context.Context, pipeline.Payload) bool {
	if s.err == nil && s.currentHeight < s.endHeight {
		s.currentHeight = s.currentHeight + 1
		return true
	}
	return false
}

func (s *backfillSource) Current() int64 {
	return s.currentHeight
}

func (s *backfillSource) Err() error {
	return s.err
}

func (s *backfillSource) Len() int64 {
	return s.endHeight - s.startHeight + 1
}

func (s *backfillSource) init() error {
	if err := s.setStartHeight(); err != nil {
		return err
	}
	if err := s.setEndHeight(); err != nil {
		return err
	}
	return nil
}

func (s *backfillSource) setStartHeight() error {
	syncable, err := s.db.Syncables.FindFirstByDifferentIndexVersion(s.sourceCfg.indexVersion)
	if err != nil {
		if err == store.ErrNotFound {
			return errors.New(fmt.Sprintf("nothing to backfill [currentIndexVersion=%d]", s.sourceCfg.indexVersion))
		}
		return err
	}

	s.currentHeight = syncable.Height
	s.startHeight = syncable.Height
	return nil
}

func (s *backfillSource) setEndHeight() error {
	syncable, err := s.db.Syncables.FindMostRecentByDifferentIndexVersion(s.sourceCfg.indexVersion)
	if err != nil {
		if err == store.ErrNotFound {
			return errors.New(fmt.Sprintf("nothing to backfill [currentIndexVersion=%d]", s.sourceCfg.indexVersion))
		}
		return err
	}

	s.endHeight = syncable.Height
	return nil
}