package syncablerepo

import (
	"fmt"
	"github.com/figment-networks/polkadothub-indexer/config"
	"github.com/figment-networks/polkadothub-indexer/models/syncable"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/utils/errors"
	"github.com/jinzhu/gorm"
)

type DbRepo interface {
	// Queries
	Exists(types.SequenceType, types.SequenceId, types.SyncableType) bool
	GetByHeight(types.SequenceType, types.SequenceId, types.SyncableType) (*syncable.Model, errors.ApplicationError)
	GetMostRecent(types.SequenceType, types.SyncableType) (*syncable.Model, errors.ApplicationError)
	GetMostRecentCommon(types.SequenceType) (*types.SequenceId, errors.ApplicationError)

	// Commands
	DbSaver
	DeletePrev(types.SequenceType, types.SequenceId) errors.ApplicationError
}

type DbSaver interface {
	Save(*syncable.Model) errors.ApplicationError
	Create(*syncable.Model) errors.ApplicationError
}

type dbRepo struct {
	client *gorm.DB
}

func NewDbRepo(c *gorm.DB) DbRepo {
	return &dbRepo{
		client: c,
	}
}

func (r *dbRepo) Exists(st types.SequenceType, sid types.SequenceId, t types.SyncableType) bool {
	q := syncable.Model{
		SequenceType: st,
		SequenceId: sid,
		Type: t,
	}
	foundTransaction := syncable.Model{}

	if err := r.client.Where(&q).First(&foundTransaction).Error; err != nil {
		return false
	}
	return true
}

func (r *dbRepo) GetByHeight(st types.SequenceType, sid types.SequenceId, t types.SyncableType) (*syncable.Model, errors.ApplicationError) {
	q := syncable.Model{
		SequenceType: st,
		SequenceId: sid,
		Type: t,
	}
	m := syncable.Model{}

	if err := r.client.Where(&q).First(&m).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errors.NewError(fmt.Sprintf("could not find syncable with query %v", q), errors.NotFoundError, err)
		}
		return nil, errors.NewError("error getting syncable by height", errors.QueryError, err)
	}
	return &m, nil
}

func (r *dbRepo) GetMostRecent(st types.SequenceType, t types.SyncableType) (*syncable.Model, errors.ApplicationError) {
	q := syncable.Model{
		SequenceType: st,
		Type: t,
	}
	m := syncable.Model{}
	if err := r.client.Debug().Where(&q).Where("processed_at IS NOT NULL").Order("sequence_id desc").First(&m).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errors.NewError("could not find most recent syncable", errors.NotFoundError, err)
		}
		return nil, errors.NewError("error getting most recent syncable", errors.QueryError, err)
	}
	return &m, nil
}

func (r *dbRepo) GetMostRecentCommon(st types.SequenceType) (*types.SequenceId, errors.ApplicationError) {
	var syncables []*syncable.Model
	for _, t := range types.Types {
		s, err := r.GetMostRecent(st, t)
		// If record is not found break immediately
		if err != nil {
			return nil, err
		}
		syncables = append(syncables, s)
	}

	// If there are not syncables yet, just start from the beginning
	if len(syncables) == 0 {
		h := types.SequenceId(config.FirstBlockHeight())
		return &h, nil
	}

	smallestH := syncables[0].SequenceId
	for _, s := range syncables {
		if s.SequenceId < smallestH {
			smallestH = s.SequenceId

		}
	}
	return &smallestH, nil
}

func (r *dbRepo) Save(m *syncable.Model) errors.ApplicationError {
	if err := r.client.Save(m).Error; err != nil {
		return errors.NewError("could not save syncable", errors.SaveError, err)
	}
	return nil
}

func (r *dbRepo) Create(m *syncable.Model) errors.ApplicationError {
	if err := r.client.Create(m).Error; err != nil {
		return errors.NewError("could not create syncable", errors.CreateError, err)
	}
	return nil
}

func (r *dbRepo) DeletePrev(st types.SequenceType, id types.SequenceId) errors.ApplicationError {
	if err := r.client.Debug().Where("sequence_type = ? AND sequence_id <= ?", st, id).Delete(&syncable.Model{}).Error; err != nil {
		return errors.NewError("could not delete syncables", errors.DeleteError, err)
	}
	return nil
}
