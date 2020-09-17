package store

import (
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/types"
)

type Syncables interface {
	BaseStore
	FindLastInSession(session int64) (syncable *model.Syncable, err error)
	FindByHeight(height int64) (syncable *model.Syncable, err error)
	FindSmallestIndexVersion() (*int64, error)
	FindMostRecent() (*model.Syncable, error)
	FindLastInEra(era int64) (syncable *model.Syncable, err error)
	FindFirstByDifferentIndexVersion(indexVersion int64) (*model.Syncable, error)
	FindMostRecentByDifferentIndexVersion(indexVersion int64) (*model.Syncable, error)
	CreateOrUpdate(val *model.Syncable) error
	SetProcessedAtForRange(reportID types.ID, startHeight int64, endHeight int64) error
}
