package extrinsicseqrepo

import (
	"fmt"
	"github.com/figment-networks/polkadothub-indexer/models/extrinsicseq"
	"github.com/figment-networks/polkadothub-indexer/models/shared"
	"github.com/figment-networks/polkadothub-indexer/types"
	"github.com/figment-networks/polkadothub-indexer/utils/errors"
	"github.com/jinzhu/gorm"
)

type DbRepo interface {
	// Queries
	Exists(types.Height) bool
	Count() (*int64, errors.ApplicationError)
	GetByHeight(types.Height) ([]extrinsicseq.Model, errors.ApplicationError)

	// Commands
	Save(*extrinsicseq.Model) errors.ApplicationError
	Create(*extrinsicseq.Model) errors.ApplicationError
}

type dbRepo struct {
	client *gorm.DB
}

func NewDbRepo(c *gorm.DB) DbRepo {
	return &dbRepo{
		client: c,
	}
}

func (r *dbRepo) Exists(h types.Height) bool {
	q := heightQuery(h)
	m := extrinsicseq.Model{}

	if err := r.client.Where(&q).First(&m).Error; err != nil {
		return false
	}
	return true
}

func (r *dbRepo) Count() (*int64, errors.ApplicationError) {
	var count int64
	if err := r.client.Table(extrinsicseq.Model{}.TableName()).Count(&count).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errors.NewError(fmt.Sprintf("could not get count of transaction sequences"), errors.NotFoundError, err)
		}
		return nil, errors.NewError("error getting count of transaction sequences", errors.QueryError, err)
	}

	return &count, nil
}

func (r *dbRepo) GetByHeight(h types.Height) ([]extrinsicseq.Model, errors.ApplicationError) {
	q := heightQuery(h)
	var ms []extrinsicseq.Model

	if err := r.client.Where(&q).Find(&ms).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errors.NewError(fmt.Sprintf("could not find transaction sequences with height %d", h), errors.NotFoundError, err)
		}
		return nil, errors.NewError("error getting transaction sequences by height", errors.QueryError, err)
	}
	return ms, nil
}

func (r *dbRepo) Save(m *extrinsicseq.Model) errors.ApplicationError {
	if err := r.client.Save(m).Error; err != nil {
		return errors.NewError("could not save transaction sequence", errors.SaveError, err)
	}
	return nil
}

func (r *dbRepo) Create(m *extrinsicseq.Model) errors.ApplicationError {
	if err := r.client.Create(m).Error; err != nil {
		return errors.NewError("could not create transaction sequence", errors.CreateError, err)
	}
	return nil
}

/*************** Private ***************/

func heightQuery(h types.Height) extrinsicseq.Model {
	return extrinsicseq.Model{
		HeightSequence: &shared.HeightSequence{
			Height: h,
		},
	}
}


