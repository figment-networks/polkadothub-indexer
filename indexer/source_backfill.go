package indexer

import (
	"context"
	"fmt"
	"github.com/figment-networks/indexing-engine/pipeline"
	"github.com/figment-networks/polkadothub-indexer/client"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/store"
	"github.com/pkg/errors"
)

var (
	_ pipeline.Source = (*backfillSource)(nil)
)

func NewBackfillSource(cfg *config.Config, db *store.Store, client *client.Client, indexVersion int64, isForLastOfSessions, isForLastOfEras bool) (*backfillSource, error) {
	src := &backfillSource{
		cfg:                 cfg,
		db:                  db,
		client:              client,
		isForLastOfSessions: isForLastOfSessions,
		isForLastOfEras:     isForLastOfEras,
		currentIndexVersion: indexVersion,
	}

	if err := src.init(); err != nil {
		return nil, err
	}

	return src, nil
}

type backfillSource struct {
	cfg                 *config.Config
	db                  *store.Store
	client              *client.Client
	isForLastOfSessions bool
	isForLastOfEras     bool
	specificHeightsMap  map[int64]int64
	currentIndexVersion int64
	currentHeight       int64
	startHeight         int64
	endHeight           int64
	err                 error
}

func (s *backfillSource) Next(context.Context, pipeline.Payload) bool {
	if s.err == nil && s.currentHeight < s.endHeight {
		s.currentHeight = s.currentHeight + 1
		return true
	}
	return false
}

func (s *backfillSource) isCurrentHeightInSpecificMap() bool {
	_, found := s.specificHeightsMap[s.currentHeight]
	return found
}

func (s *backfillSource) HasOnlySpecificHeightsToRunStages() bool {
	return s.isForLastOfSessions || s.isForLastOfEras
}

func (s *backfillSource) Current() int64 {
	return s.currentHeight
}

func (s *backfillSource) Err() error {
	return s.err
}

func (s *backfillSource) SkipRunningStagesForCurrentHeight() bool {
	if s.HasOnlySpecificHeightsToRunStages() && !s.isCurrentHeightInSpecificMap() {
		return true
	}
	return false
}

func (s *backfillSource) Len() int64 {
	return s.endHeight - s.startHeight + 1
}

func (s *backfillSource) init() error {
	if err := s.setSpecificHeights(); err != nil {
		return err
	}
	if err := s.setStartHeight(); err != nil {
		return err
	}
	if err := s.setEndHeight(); err != nil {
		return err
	}
	return nil
}

func (s *backfillSource) setStartHeight() error {
	syncable, err := s.db.Syncables.FindFirstByDifferentIndexVersion(s.currentIndexVersion)
	if err != nil {
		if err == store.ErrNotFound {
			return errors.New(fmt.Sprintf("nothing to backfill [currentIndexVersion=%d]", s.currentIndexVersion))
		}
		return err
	}

	s.currentHeight = syncable.Height
	s.startHeight = syncable.Height
	return nil
}

func (s *backfillSource) setEndHeight() error {
	syncable, err := s.db.Syncables.FindMostRecentByDifferentIndexVersion(s.currentIndexVersion)
	if err != nil {
		if err == store.ErrNotFound {
			return errors.New(fmt.Sprintf("nothing to backfill [currentIndexVersion=%d]", s.currentIndexVersion))
		}
		return err
	}

	s.endHeight = syncable.Height
	return nil
}

func (s *backfillSource) setSpecificHeights() error {
	if s.HasOnlySpecificHeightsToRunStages() {
		syncables, err := s.db.Syncables.FindAllByLastInSessionOrEra(s.currentIndexVersion, s.isForLastOfSessions, s.isForLastOfEras)
		if err != nil {
			if err == store.ErrNotFound {
				return errors.New(fmt.Sprintf("no specific heights to backfill [currentIndexVersion=%d]", s.currentIndexVersion))
			}
			return err
		}
		s.generateMapForSpecificHeights(syncables)
	}
	return nil
}

func (s *backfillSource) generateMapForSpecificHeights(syncables []model.Syncable) {
	specificHeightsMap := make(map[int64]int64)
	for i := 0; i < len(syncables); i++ {
		specificHeightsMap[syncables[i].Height] = syncables[i].Height
	}
	s.specificHeightsMap = specificHeightsMap
}
