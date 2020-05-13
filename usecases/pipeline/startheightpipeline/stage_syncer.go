package startheightpipeline

import (
	"context"
	"github.com/figment-networks/polkadothub-indexer/models/syncable"
	"github.com/figment-networks/polkadothub-indexer/repos/syncablerepo"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/utils/errors"
	"github.com/figment-networks/polkadothub-indexer/utils/pipeline"
)

type Syncer interface {
	Process(context.Context, pipeline.Payload) (pipeline.Payload, error)
}

type syncer struct {
	syncableDbRepo    syncablerepo.DbRepo
	syncableProxyRepo syncablerepo.ProxyRepo
}

func NewSyncer(syncableDbRepo syncablerepo.DbRepo, syncableProxyRepo syncablerepo.ProxyRepo) Syncer {
	return &syncer{
		syncableDbRepo:    syncableDbRepo,
		syncableProxyRepo: syncableProxyRepo,
	}
}

func (s *syncer) Process(ctx context.Context, p pipeline.Payload) (pipeline.Payload, error) {
	payload := p.(*payload)
	h := payload.CurrentHeight

	// Sync block
	b, err := s.sync(types.SyncableTypeBlock, h)
	if err != nil {
		return nil, err
	}
	payload.BlockSyncable = b

	return payload, nil
}

/*************** Private ***************/

func (s *syncer) sync(t types.SyncableType, h types.Height) (*syncable.Model, errors.ApplicationError) {
	existingSyncable, err := s.syncableDbRepo.GetByHeight(types.SequenceTypeHeight, types.SequenceId(h), t)
	if err != nil {
		if err.Status() == errors.NotFoundError {
			// Create if not found
			m, err := s.syncableProxyRepo.GetByHeight(t, h)
			if err != nil {
				return nil, err
			}

			err = s.syncableDbRepo.Create(m)
			if err != nil {
				return nil, err
			}

			return m, nil
		}
	}
	return existingSyncable, nil
}
