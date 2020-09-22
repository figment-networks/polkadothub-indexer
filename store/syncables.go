package store

import (
	"github.com/figment-networks/polkadothub-indexer/model"
	"github.com/figment-networks/polkadothub-indexer/types"
)

type Syncables interface {
	BaseStore
	CreateOrUpdate(val *model.Syncable) error
	FindByHeight(height int64) (syncable *model.Syncable, err error)
	FindFirstByDifferentIndexVersion(indexVersion int64) (*model.Syncable, error)
	FindLastInEra(era int64) (syncable *model.Syncable, err error)
	FindLastInSession(session int64) (syncable *model.Syncable, err error)
	FindMostRecent() (*model.Syncable, error)
	FindMostRecentByDifferentIndexVersion(indexVersion int64) (*model.Syncable, error)
	FindSmallestIndexVersion() (*int64, error)
	SetProcessedAtForRange(reportID types.ID, startHeight int64, endHeight int64) error
}
