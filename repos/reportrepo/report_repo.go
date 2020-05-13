package reportrepo

import (
	"github.com/figment-networks/polkadothub-indexer/models/report"
	"github.com/figment-networks/polkadothub-indexer/utils/errors"
	"github.com/jinzhu/gorm"
)

type DbRepo interface {
	// Commands
	Create(*report.Model) errors.ApplicationError
	Save(*report.Model) errors.ApplicationError
}

type dbRepo struct{
	client *gorm.DB
}

func NewDbRepo(c *gorm.DB) DbRepo {
	return &dbRepo{
		client: c,
	}
}

func (r *dbRepo) Create(m *report.Model) errors.ApplicationError {
	if err := r.client.Create(m).Error; err != nil {
		return errors.NewError("could not create block", errors.CreateError, err)
	}
	return nil
}

func (r *dbRepo) Save(m *report.Model) errors.ApplicationError {
	if err := r.client.Save(m).Error; err != nil {
		return errors.NewError("could not save report", errors.CreateError, err)
	}
	return nil
}

