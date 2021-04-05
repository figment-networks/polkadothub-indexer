package store

import (
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/types"
)

type syncables interface {
	CreateOrUpdate(val *model.Syncable) error
	FindByHeight(height int64) (syncable *model.Syncable, err error)
	FindFirstByDifferentIndexVersion(indexVersion int64) (*model.Syncable, error)
	FindLastInEra(era int64) (syncable *model.Syncable, err error)
	FindLastEndOfEra() (syncable *model.Syncable, err error)
	FindLastInSession(session int64) (syncable *model.Syncable, err error)
	FindLastEndOfSession() (syncable *model.Syncable, err error)
	FindLastInSessionForHeight(height int64) (syncable *model.Syncable, err error)
	FindMostRecentByDifferentIndexVersion(indexVersion int64) (*model.Syncable, error)
	FindSmallestIndexVersion() (*int64, error)
	SaveSyncable(*model.Syncable) error
	SetProcessedAtForRange(reportID types.ID, startHeight int64, endHeight int64) error
	FindAllByLastInSessionOrEra(indexVersion int64, isLastInSession, isLastInEra bool, start, end int64) ([]model.Syncable, error)
}

type FindMostRecenter interface {
	FindMostRecent() (*model.Syncable, error)
}
