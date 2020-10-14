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

func NewBackfillSource(cfg *config.Config, db *store.Store, client *client.Client, indexVersion int64) (*backfillSource, error) {
	src := &backfillSource{
		cfg:    cfg,
		db:     db,
		client: client,
		currentIndexVersion: indexVersion,
	}

	if err := src.setHeightValues(); err != nil {
		return nil, err
	}

	return src, nil
}

type backfillSource struct {
	cfg    *config.Config
	db     *store.Store
	client *client.Client

	currentIndexVersion int64
	endSyncsOfSessionsAndEras []int64
	length  int64
	currentHeight int64
	endHeight     int64
	err           error
}

func (s *backfillSource) Next(context.Context, pipeline.Payload) bool {
	if s.err == nil && s.currentHeight < s.endHeight {
		s.endSyncsOfSessionsAndEras = s.endSyncsOfSessionsAndEras[1:]
		s.setCurrentHeight()
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

func (s *backfillSource) setHeightValues() error {
	if err := s.setEndSyncsOfSessionsAndEras(); err != nil {
		return err
	}
	s.setCurrentHeight()
	s.setEndHeight()
	return nil
}

func (s *backfillSource) setEndSyncsOfSessionsAndEras() error {
	endSyncsOfSessionsAndEras, err := s.db.Syncables.FindEndSyncsOfSessionsAndEras(s.currentIndexVersion)
	if err != nil {
		if err == store.ErrNotFound {
			return errors.New(fmt.Sprintf("nothing to backfill [currentIndexVersion=%d]", s.currentIndexVersion))
		}
		return err
	}

	var heights []int64
	for i := 0; i < len(endSyncsOfSessionsAndEras); i++ {
		heights = append(heights, endSyncsOfSessionsAndEras[i].Height)
	}
	s.endSyncsOfSessionsAndEras = heights
	s.length = int64(len(heights))
	return nil
}

func (s *backfillSource) setCurrentHeight(){
	s.currentHeight =  s.endSyncsOfSessionsAndEras[0]
}

func (s *backfillSource) setEndHeight(){
	s.endHeight = s.endSyncsOfSessionsAndEras[len(s.endSyncsOfSessionsAndEras)-1]
}
